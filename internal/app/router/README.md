# router system

English

```bash
┌─────────────────────────┐
│      Router Registry    │
│   RouteRegistry         │
├─────────┬───────┬───────┤
│ Global  │ v1     │ v2     │
└─────────┴───┬───┴───────┘
              │
    ┌─────────┴──────────┐
    │     Version Router │
    │  VersionRouter     │
    ├──────────┬─────────┤
    │ Public   │ Private  │
    └──────────┴─────────┘
```

中文

```bash
┌─────────────────────────┐
│     路由注册表          │
│   RouteRegistry         │
├─────────┬───────┬───────┤
│ 全局路由 │ v1版本│ v2版本│
└─────────┴───┬───┴───────┘
              │
    ┌─────────┴──────────┐
    │    版本路由器      │
    │  VersionRouter     │
    ├──────────┬─────────┤
    │ 公共路由  │ 私有路由 │
    └──────────┴─────────┘
```
