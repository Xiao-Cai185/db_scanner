# DB_SCANNER

`db_scanner` 是一个简单的命令行工具，基于[SecurityPaper/thai_bone](https://github.com/SecurityPaper/thai_bone/)二次开发，用于扫描数据库中的敏感信息，并通过正则表达式匹配识别常见的敏感数据类型。它支持多种操作模式，包括数据库列表显示、特定表查询和详细输出模式等，适合信息安全分析与隐私数据筛查。

## 功能简介

- **扫描数据库中的敏感数据**：支持多种类型的敏感信息，如电子邮件、身份证号码、电话号码和 IP 地址等。
- **支持不同查询模式**：
  - 显示数据库列表（通过 `--showdbs` 参数）
  - 查询特定表（通过 `--table` 参数）
  - 启用详细模式显示完整的表数据（通过 `-v` 参数）
- **支持正则表达式匹配**：可针对表中的字段使用特定规则进行敏感信息识别。（规则基于`ruler.ymal`）

## 使用方法

### 基本命令

```bash
db_scanner [选项] [参数]
```

---

## **用法**

| 参数                  | 描述                               |
|-----------------------|------------------------------------|
| `--host, -h`          | 数据库主机地址                     |
| `--user, -u`          | 数据库用户名                       |
| `--pass, -p`          | 数据库密码                         |
| `--port, -P`          | 数据库端口号（默认：3306）         |
| `--char, -c`          | 字符编码（默认：utf8mb4）          |
| `--database, --db`    | 数据库名                           |
| `--verbose, -v`       | 是否显示符合条件的每一行表数据     |
| `--table, -t`         | 指定要查询的表名                   |
| `--showdbs, -s`       | 展示所有数据库名称                 |
| `--help`              | 显示帮助信息                       |

## **例子**

### 1. 查询数据库列表

```bash
db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 -s
```
```bash
db_scanner -h 127.0.0.1 -u root -p password -s
```

### 2. 查询特定数据库

```bash
db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 --db testdb
```
```bash
db_scanner -h 127.0.0.1 -u root -p password --db testdb
```

### 3. 查询特定表

```bash
db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 --db testdb -t users
```
```bash
db_scanner -h 127.0.0.1 -u root -p password --db testdb -t users
```

### 4. 查询特定表并显示所有行数据

```bash
db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 --db testdb -t users -v
```
```bash
db_scanner -h 127.0.0.1 -u root -p password --db testdb -t users -v
```

---

## 编译指南

`db_scanner` 支持在多平台上运行，包括 Windows、Linux 和 macOS。可以使用 Go 的交叉编译功能生成不同系统的可执行文件。

### Windows 编译

```bash
go build -o db_scanner_windows.exe
```

### Linux 编译

```bash
GOOS=linux GOARCH=amd64 go build -o db_scanner_linux
```

### macOS ARM64 编译

```bash
GOOS=darwin GOARCH=arm64 go build -o db_scanner_macos_arm64
```

### 运行编译后的文件

- Windows:
  ```powershell
  .\db_scanner_windows.exe --help
  ```
- Linux:
  ```bash
  ./db_scanner_linux --help
  ```
- macOS:
  ```bash
  ./db_scanner_macos_arm64 --help
  ```

以下是根据工具用法和例子整理的“常见问题”板块：

---

## **常见问题**

### 1. 如何查询所有数据库的名称？

使用 `--showdbs` 参数，可以快速显示数据库列表：
```bash
db_scanner -h 127.0.0.1 -u root -p password -s
```
此命令会以表格格式显示所有数据库的名称，适用于在不知道数据库名称的情况下进行查询。

### 2. 如果我只想查询某个特定数据库，应该怎么做？

你可以使用 `--db` 参数指定目标数据库：
```bash
db_scanner -h 127.0.0.1 -u root -p password --db testdb
```
这会查询并显示 `testdb` 数据库中的敏感信息。

### 3. 如何查询特定表中的数据？

使用 `--table` 参数来指定你要查询的表名。例如，查询 `testdb` 数据库中的 `users` 表：
```bash
db_scanner -h 127.0.0.1 -u root -p password --db testdb -t users
```
该命令会显示该表中的第一行数据。如果你想显示所有行，可以加上 `-v` 参数。

### 4. 如何显示特定表中的所有数据？

你可以在 `--table` 参数的基础上添加 `-v` 参数，显示符合条件的每一行表数据：
```bash
db_scanner -h 127.0.0.1 -u root -p password --db testdb -t users -v
```
这样可以看到 `users` 表中的所有行数据。

### 5. 如果我不确定字符编码或者端口号，该怎么办？

字符编码默认为 `utf8mb4`，端口号默认为 `3306`。如果需要更改，可以通过 `--char` 和 `--port` 参数指定新的值，例如：
```bash
db_scanner -h 127.0.0.1 -u root -p password -P 3307 -c latin1 --db testdb
```
在不指定这些参数的情况下，工具会使用默认值。

---


## 许可证

此项目遵循 [MIT License](LICENSE)。

---

![image](https://github.com/user-attachments/assets/0be29c3f-874f-4cfe-9c5c-0994170d9e00)

