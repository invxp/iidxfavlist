# IIDX Favorite List Editor

### 快速开始
```
$ go get github.com/invxp/iidxfavlist
$ cd iidxfavlist
$ go build cmd/iidxfavlist
$ ./iidxfavlist
```
### beatmaniaIIDX 歌单编辑器
1. 能够自定义编辑歌单
2. 能够方便的模糊查找歌单与游戏系统库的歌曲
3. 本应用依赖playlist插件

### 使用教程
#### 请确保放到beatmaniaIIDX游戏目录下
#### 输入对应的命令即可
* e: 增加/修改/删除歌单
* l: 查询歌单
* r: 重命名歌单
* s: 搜索游戏歌曲库.如:'s {id}/{artist}/{songname}' 除ID外支持模糊查询
* f: 搜索当前歌单库.如:'f {id}/{artist}/{songname}' 除ID外支持模糊查询
* b: 返回上一级菜单(部分功能具备)
* q: 退出应用
### 最佳实践
