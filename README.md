### 由于现在三国策略游戏爆火，比如刷土之滨、三国志等等，东南亚大部分依赖国外的游戏 ，电脑渐渐成为新兴市场，提出了无敌三国策略游戏。
### 此项目包括 websocket 框架的搭建，登录注册、建筑升级、部队出征、联盟管理和聊天服等等，项目主要分为网关服务、游戏服务、聊天服务，登录认证服务，web服务。
### 游戏通过websocket协议将加密信息传递到网关websocket服务，然后网关将其转发为具体的业务上，项目大量使用了go+channel的方式用以实时更新资源和实时数据，游戏数据以缓存的形式，加快访问，使用了锁保证并发安全，根据游戏服务器，数据库做了分库处理，不同服务器的用户数据，会落在不同的库上，日志数据存在mongo集群上。
### 项目自定义了一套websocket框架，实现负载均衡的路由，无状态服务随机分配，有状态服务根据状态标识路由，还有请求响应统一封装，上下文支持，中间件支持，实现统一日志中间件，认证中间件；对于大规模战斗，采用开启大量协程提供单台服务器性能，简化战斗逻辑的方式应对。