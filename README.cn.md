# DuckGo: 轻松创建 Go/XGo 版 DuckDB UDF

[![Go Reference](https://pkg.go.dev/badge/github.com/ma6174/duckgo.svg)](https://pkg.go.dev/github.com/ma6174/duckgo)

`DuckGo` 是一个 Go 语言库，旨在极大地简化为 [DuckDB](https://duckdb.org/) 创建用户定义函数 (UDF) 的过程。它通过 Go 的反射机制，让您可以直接将一个普通的 Go 函数无缝转换为 DuckDB 的标量函数 (Scalar UDF)，同时通过[ixgo](https://github.com/goplus/ixgo)支持从 Go/XGo 脚本动态创建加载 UDF，无需预先编译。

## 核心功能

- **从原生 Go 函数创建**: 直接将你的 Go 函数（例如 `func(a, b int) int`) 转换为 DuckDB UDF。
- **自动类型映射**: 自动处理 Go 类型与 DuckDB 类型之间的转换，支持多种数据类型。
- **支持可变参数**: 支持 Go 的可变参数函数 (variadic functions)。
- **动态脚本加载**: 无需编译！可以直接从 `.go`、`.xgo` 源文件动态加载函数作为 UDF。
- **SQL 内直接加载**: 提供一个辅助函数，允许您直接在 SQL 查询中加载和注册脚本中的 UDF。
- **错误处理**: 妥善处理 UDF 执行过程中的 `panic`，并将其转换为 DuckDB 错误返回。

## 安装

```bash
go get github.com/ma6174/duckgo
```

## 重要提示：关于动态脚本功能

本项目的动态脚本加载功能（由 `script` 包提供）依赖于 [ixgo](https://github.com/goplus/ixgo)。由于 `ixgo` 的工作原理，当您编译或运行任何使用了 `script` 包的代码时，**必须**添加一个特定的链接器标志来禁用符号链接检查。

否则，您会遇到链接器错误。

**编译时:**
```bash
go build -ldflags="-checklinkname=0" .
```

**运行时:**
```bash
go run -ldflags="-checklinkname=0" .
```

如果您在自己的项目里引用了本库的 `script` 包，请务必在您的编译和运行命令中加入此标志。

## 使用方法

下面是几个例子，展示了如何使用 `DuckGo`。完整的可运行示例代码存放在 [`example`](./example) 目录下。

### 示例 1: 从原生 Go 函数创建 UDF

这是最基本的使用方式。任何普通的 Go 函数都可以被注册。

核心逻辑是使用 `udf.BuildScalarUDF` 将原生 Go 函数包装起来，然后通过 `duckdb.RegisterScalarUDF` 进行注册。

**完整代码请参见: [`example/simple_udf/main.go`](./example/simple_udf/main.go)**

### 示例 2: 从 Go 脚本动态加载 UDF

这是 `DuckGo` 最强大的功能之一。您无需编译，可以直接将一个 Go 源文件中的函数注册为 UDF。

通过 `script.AddIXGoUDFFromFile` 函数，您可以指定一个 `.go` 文件和需要加载的函数名。

**完整代码请参见: [`example/script_udf/`](./example/script_udf/)**

### 示例 3: 通过 SQL 直接加载脚本 UDF

为了让动态加载更加灵活，您甚至可以直接在 SQL 中调用一个辅助函数来加载 UDF。

首先调用 `script.EnableRegisterUDFFromSQL(db)` 启用该功能，之后就可以在 SQL 中使用 `add_ixgo_udf` 函数了。

**完整代码请参见: [`example/sql_load_udf/`](./example/sql_load_udf/)**

### 示例 4：导入外部包

`DuckGo` 支持从外部包（如 `github.com/goccy/go-json`）导入函数作为 UDF。这通过使用 `go generate` 和 [qexp](https://github.com/goplus/ixgo/tree/main/cmd/qexp) 预处理外部包来实现。

**完整代码请参见: [`example/external_package/`](./example/external_package/)**

## 包概览

- **`udf`**: 核心包，负责将原生的 Go 函数转换为 DuckDB UDF。
- **`script`**: 提供从 Go/XGo 脚本动态加载 UDF 的功能。

## 许可证

本项目采用 [MIT](LICENSE) 许可证。