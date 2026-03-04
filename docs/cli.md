# 命令行参数

```
./deeptrace [options]
```

| 参数 | 类型 | 默认值 | 说明 |
|---|---|---|---|
| `-configs` | string | `conf.d` | 配置目录路径 |
| `-interval` | int | `0` | 全局采集间隔（秒），覆盖配置文件中的 `global.interval` |
| `-plugins` | string | `""` | 只运行指定插件，多个用 `:` 分隔，如 `disk:procnum` |
| `-loglevel` | string | `""` | 日志级别，覆盖配置文件，可选 `debug` `info` `warn` `error` `fatal` |
| `-version` | bool | `false` | 显示版本号 |

## 常用场景

### 测试模式


```bash
./deeptrace -test
```

### 只运行指定插件

```bash
./deeptrace -test -plugins disk:procnum
```

### 指定配置目录

```bash
./deeptrace -configs /etc/deeptrace/conf.d
```

### Windows 服务管理

```bash
deeptrace.exe -win-service-install
deeptrace.exe -win-service-start
deeptrace.exe -win-service-stop
deeptrace.exe -win-service-uninstall
```
