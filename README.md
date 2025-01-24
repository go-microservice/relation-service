# relation service

好友关系服务

## Feature

- 关注/取消关注
- 查询用户关注列表
  - 本地缓存->redis缓存->数据库->zset
  - 如果是大V(指定大小的粉丝数)需要缓存到本地缓存
- 查询用户粉丝列表
  - 最近的10000个粉丝，查询redis, 查不到再查数据库
- 查询用户的关注数与粉丝数
- 查询用户关注关系
  - 单个查询: 用户A是关注了用户B, 用户B是否关注了用户A, 是否相互关注
  - 批量查询关注: 用户A是否关注了B,C,D...
  - 批量查询粉丝: 用户B,C,D...是否关注了A

## 数据库表设计

```sql
-- 关注表
CREATE TABLE `user_following` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '发起关注的人',
  `followed_uid` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '被关注用户的uid',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '关注状态 1:已关注 0:取消关注',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_uid_fuid` (`user_id`,`followed_uid`),
  KEY `idx_following` (`user_id`,`followed_uid`,`status`),
  KEY `idx_following_list` (`user_id`,`status`,`updated_at`,`followed_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户关注表';

-- 粉丝表
CREATE TABLE `user_follower` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '用户id',
  `follower_uid` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '粉丝的uid',
  `status` tinyint(1) unsigned NOT NULL DEFAULT '0' COMMENT '状态 1:已关注 0:取消关注',
  `created_at` datetime DEFAULT NULL,
  `updated_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_uid_fid` (`user_id`,`follower_uid`),
  -- KEY `idx_follower` (`user_id`,`follower_uid`), 与上面重复可删除
  KEY `idx_follower_list` (`user_id`,`status`,`updated_at`,`follower_uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='用户粉丝表';
```

## 关键SQL语句

```sql
-- 用户A是否关注了用户B
SELECT followed_uid FROM user_following WHERE user_id=用户A AND followed_uid=用户B AND status=1;
-- 用户A的关注列表
SELECT followed_uid FROM user_following WHERE user_id=用户A AND status=1 ORDER BY updated_at DESC;
-- 批量查询用户A是否关注了用户B,C,D
SELECT followed_uid FROM user_following WHERE user_id=用户A AND followed_uid IN(用户B, 用户C, 用户D);

-- 查询用户A粉丝列表
SELECT follower_uid FROM user_follower WHERE user_id=用户A AND status=1 ORDER BY updated_at DESC;
-- 批量查询用户B,C,D是否是用户A的粉丝
SELECT follower_uid FROM user_follower WHERE user_id=用户A AND follower_uid IN(用户B, 用户C, 用户D);
```

## 缓存处理

主要使用redis和本地缓存， 缓存的内容主要包含如下:

- 关注列表 ZSET (全量缓存)
  - 更新策略: 先更新数据库再删除缓存
  - 也可以和粉丝列表一样的更新策略来更新
  - 大V的关注列表-写入本地缓存
    - 特点: 访问量大、数据不易变
    - 好处: 减少访问redis, 提高访问性能
- 粉丝列表 ZSET（缓存最近的10000个粉丝）
  - 更新策略: 先更新数据库再更新缓存，提高缓存命中率
  - 更新方式：同步修改，或者定于following 的binglog异步更新缓存
  - 原因: 频繁的发生关注事件会造成频繁的删除和创建缓存
  - 大V的粉丝列表，特点: 粉丝变动频繁，不适合使用本地缓存
  - 普通用户，特点: 粉丝变动较小，粉丝列表访问量也小，使用本地缓存后缓存命中率不高，也不适合使用本地缓存
  - 小于10000， 查zset,查db
  - 大于10000，粉丝列表的zset可能无数据，查hash对象缓存,查到则返回，查不到回源数据库，再写入hash

## 使用场景

- 单个用户关注关系查询（查询关注列表缓存）
- 批量用户关注关系查询（查询关注列表缓存）
- 批量查询粉丝关系

## 注意事项

- 限制一个用户可以关注的最大人数，比如2000，防止用户利用关注刷粉、刷流量，以及影响用户体验
- 由于有关注人数限制，所以关注列表可以全量缓存到redis
- 粉丝列表没有限制
  - 由于粉丝列表太多，所以只会缓存最近的比如10000个粉丝(可以满足大部分场景，也可以应对高并发的请求)
  - 近10000个的粉丝优先查询redis, 大于10000的再查数据库
  - 为防止大量请求超过10000的粉丝列表，进而访问到数据库可能造成的宕机，所以这里需要进行适当的限流处理

## 编译及运行

```bash
# 编译
go build

# 运行
./relation-service -c=config -e=dev
```
