package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"time" 
	"io" 
	"path/filepath"


	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"github.com/fatih/color"
)

var symbols map[string]func(a ...interface{}) string

func init() {
    // 初始化符号与颜色的映射关系
    symbols = map[string]func(a ...interface{}) string{
        "[+]": color.New(color.FgGreen).SprintFunc(),  // 绿色标识符
        "[-]": color.New(color.FgRed).SprintFunc(),    // 红色标识符
        "[!]": color.New(color.FgYellow).SprintFunc(), // 黄色标识符
        "[*]": color.New(color.FgBlue).SprintFunc(),   // 蓝色标识符
    }
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// 过滤函数，用于移除颜色标识符
func stripANSI(input string) string {
    return ansiRegex.ReplaceAllString(input, "")
}

type FilterWriter struct {
    Writer    io.Writer
    StripFunc func(string) string
}

func (fw *FilterWriter) Write(p []byte) (n int, err error) {
    // 只将过滤后的内容写入文件
    strippedText := fw.StripFunc(string(p))
    n, err = fw.Writer.Write([]byte(strippedText))
    return n, err
}

// 修改函数签名，添加 multiWriter 参数
func displayLogo(multiWriter io.Writer) {
	fmt.Fprintf(multiWriter, `
 _____     ______     ______     ______     ______     __   __     __   __     ______     ______
/\  __-.  /\  == \   /\  ___\   /\  ___\   /\  __ \   /\ "-.\ \   /\ "-.\ \   /\  ___\   /\  == \
\ \ \/\ \ \ \  __<   \ \___  \  \ \ \____  \ \  __ \  \ \ \-.  \  \ \ \-.  \  \ \  __\   \ \  __<
 \ \____-  \ \_____\  \/\_____\  \ \_____\  \ \_\ \_\  \ \_\\"\_\  \ \_\\"\_\  \ \_____\  \ \_\ \_\
  \/____/   \/_____/   \/_____/   \/_____/   \/_/\/_/   \/_/ \/_/   \/_/ \/_/   \/_____/   \/_/ /_/

   Ver:2.3 Rebuild:Xiao_Cai185
	
`)
}

func isHan(r rune) bool {
	return unicode.Is(unicode.Han, r)
}

// 优化回显
func printTableHeader(columns []string, multiWriter io.Writer) {
	fmt.Fprint(multiWriter, "+")
	for range columns {
		fmt.Fprint(multiWriter, "-------------------------+")
	}
	fmt.Fprintln(multiWriter)
	fmt.Fprint(multiWriter, "|")
	for _, col := range columns {
		fmt.Fprintf(multiWriter, " %-23s |", col)
	}
	fmt.Fprintln(multiWriter)
	fmt.Fprint(multiWriter, "+")
	for range columns {
		fmt.Fprint(multiWriter, "-------------------------+")
	}
	fmt.Fprintln(multiWriter)
}

// 优化回显
func printTableRow(data map[string]interface{}, columns []string, multiWriter io.Writer) {
	fmt.Fprint(multiWriter, "|")
	for _, col := range columns {
		val, ok := data[col]
		if ok {
			switch v := val.(type) {
			case string:
				fmt.Fprintf(multiWriter, " %-23s |", v)
			case int, int32, int64:
				fmt.Fprintf(multiWriter, " %-23d |", v)
			case float64:
				fmt.Fprintf(multiWriter, " %-23.2f |", v)
			default:
				fmt.Fprintf(multiWriter, " %-23v |", v)
			}
		} else {
			fmt.Fprintf(multiWriter, " %-23s |", "")
		}
	}
	fmt.Fprintln(multiWriter)
	fmt.Fprint(multiWriter, "+")
	for range columns {
		fmt.Fprint(multiWriter, "-------------------------+")
	}
	fmt.Fprintln(multiWriter)
}

// 修改函数签名，添加 multiWriter 参数
func printDatabaseList(databases []string, multiWriter io.Writer) {
	fmt.Fprint(multiWriter, "+-------------------------+\n")
	fmt.Fprint(multiWriter, "| Database Name           |\n")
	fmt.Fprint(multiWriter, "+-------------------------+\n")
	for _, dbName := range databases {
		fmt.Fprintf(multiWriter, "| %-23s |\n", dbName)
	}
	fmt.Fprint(multiWriter, "+-------------------------+\n")
}

func main() {
	// 定义命令行参数
	host := flag.String("host", "", "数据库主机地址")
	user := flag.String("user", "", "数据库用户名")
	pass := flag.String("pass", "", "数据库密码")
	port := flag.String("port", "3306", "数据库端口号，默认3306")
	char := flag.String("char", "utf8mb4", "字符编码，默认utf8mb4")
	database := flag.String("database", "", "数据库名")
	verbose := flag.Bool("verbose", false, "是否显示符合条件的每一行表数据")
	help := flag.Bool("help", false, "显示帮助信息")
	table := flag.String("table", "", "指定要查询的表名")
	showdbs := flag.Bool("showdbs", false, "展示所有数据库名称")
	outputFile := flag.String("output", "", "指定输出的文件名")

	// 短命令别名
	flag.StringVar(host, "h", "", "数据库主机地址")
	flag.StringVar(user, "u", "", "数据库用户名")
	flag.StringVar(pass, "p", "", "数据库密码")
	flag.StringVar(port, "P", "3306", "数据库端口号，默认3306")
	flag.StringVar(char, "c", "utf8mb4", "字符编码，默认utf8mb4")
	flag.StringVar(database, "db", "", "数据库名")
	flag.BoolVar(verbose, "v", false, "是否显示符合条件的每一行表数据")
	flag.StringVar(table, "t", "", "指定要查询的表名")
	flag.BoolVar(showdbs, "s", false, "展示所有数据库名称")
	flag.StringVar(outputFile, "o", "", "输出回显txt的文件名称") 

	// 解析命令行参数
	flag.Parse()

	// 创建捕获输出文件
	var outputWriter *os.File
	if *outputFile != "" {
		var err error
		outputWriter, err = os.OpenFile(*outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("%s 无法打开文件 %s 进行输出: %v\n", symbols["[-]"]("[-]"), *outputFile, err)
			os.Exit(1)
		}
		defer outputWriter.Close()

		// 获取当前时间并写入文件
		currentTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(outputWriter, "报告生成时间: %s\n\n", currentTime)
	}

	// 定义一个多路输出到终端和文件的Writer
    multiWriter := io.MultiWriter(os.Stdout)
    if *outputFile != "" {
        multiWriter = io.MultiWriter(os.Stdout, outputWriter)
		
		// 包装multiWriter，终端输出保留颜色，文件输出过滤掉ANSI标识符
		multiWriter = &FilterWriter{
            Writer:    multiWriter,
            StripFunc: stripANSI,
        }
	}

	// 使用 multiWriter 输出
	displayLogo(multiWriter)

	// 如果没有传入任何参数，输出帮助信息并退出
	programName := filepath.Base(os.Args[0])
	if len(os.Args) == 1 || *help {
        fmt.Fprintln(multiWriter, "用法:")
        fmt.Fprintf(multiWriter, "  --host, -h         数据库主机地址\n")
        fmt.Fprintf(multiWriter, "  --user, -u         数据库用户名\n")
        fmt.Fprintf(multiWriter, "  --pass, -p         数据库密码\n")
        fmt.Fprintf(multiWriter, "  --port, -P         数据库端口号（默认：3306）\n")
        fmt.Fprintf(multiWriter, "  --char, -c         字符编码（默认：utf8mb4）\n")
        fmt.Fprintf(multiWriter, "  --database, --db   数据库名\n")
        fmt.Fprintf(multiWriter, "  --verbose, -v      是否显示表每一行数据\n")
        fmt.Fprintf(multiWriter, "  --table, -t        指定要查询的表名\n")
        fmt.Fprintf(multiWriter, "  --showdbs, -s      展示所有数据库名称\n")
        fmt.Fprintf(multiWriter, "  --output, -o       导出扫描报告到本地文件\n")
        fmt.Fprintf(multiWriter, "  --help             显示帮助信息\n")
        fmt.Fprintln(multiWriter, "例子:")
        fmt.Fprintf(multiWriter, "  1. 查询数据库列表:\n")
        fmt.Fprintf(multiWriter, "     %s -h 127.0.0.1 -u root -p password -s\n", programName)
        fmt.Fprintf(multiWriter, "  2. 查询特定数据库:\n")
        fmt.Fprintf(multiWriter, "     %s -h 127.0.0.1 -u root -p password --db testdb\n", programName)
        fmt.Fprintf(multiWriter, "  3. 查询特定表:\n")
        fmt.Fprintf(multiWriter, "     %s -h 127.0.0.1 -u root -p password --db testdb -t users\n", programName)
        fmt.Fprintf(multiWriter, "  4. 查询特定表并显示所有行数据:\n")
        fmt.Fprintf(multiWriter, "     %s -h 127.0.0.1 -u root -p password --db testdb -t users -v\n", programName)
        fmt.Fprintf(multiWriter, "  5. 导出扫描报告到本地文件:\n")
        fmt.Fprintf(multiWriter, "     %s -h 127.0.0.1 -u root -p password --db testdb -t users -v --output report.txt\n", programName)
        os.Exit(0)
    }
	

	// 如果指定了 --showdbs 参数，则仅展示数据库列表
	if *showdbs {
		if *host == "" || *user == "" || *pass == "" {
			fmt.Fprintln(multiWriter, "错误：必须提供数据库主机地址、用户名和密码。")
			fmt.Fprintln(multiWriter, "使用 --help 获取更多信息。")
			os.Exit(1)
		}

		mysqlpath := strings.Join([]string{*user, ":", *pass, "@tcp(", *host, ":", *port, ")/", "?charset=", *char}, "")

		db, err := gorm.Open(mysql.Open(mysqlpath), &gorm.Config{})
		if err != nil {
			fmt.Fprintf(multiWriter, "数据库连接失败: %s\n", err)
			os.Exit(1)
		}

		var databases []map[string]interface{}
		db.Raw("SHOW DATABASES").Scan(&databases)

		dbNames := make([]string, len(databases))
		for i, dbName := range databases {
			for _, name := range dbName {
				dbNames[i] = name.(string)
			}
		}

		printDatabaseList(dbNames, multiWriter) // 传递 multiWriter
		os.Exit(0)
	}

	// 检查必须的参数是否缺失
	if *host == "" || *user == "" || *pass == "" || *database == "" {
		fmt.Fprintln(multiWriter, "错误：必须提供数据库主机地址、用户名、密码和数据库名。")
		fmt.Fprintln(multiWriter, "使用 --help 获取更多信息。")
		os.Exit(1)
	}

	// 加载规则文件
	viper.SetConfigName("ruler")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Fprintf(multiWriter, "%s 无法读取规则文件: %s\n", symbols["[-]"]("[-]"), err)
		os.Exit(1)
	}
	ruler := viper.GetStringMapString("ruler")

	// 拼接MySQL连接字符串
	mysqlpath := strings.Join([]string{*user, ":", *pass, "@tcp(", *host, ":", *port, ")/", *database, "?charset=", *char}, "")

	db, err := gorm.Open(mysql.Open(mysqlpath), &gorm.Config{}) // 链接数据库
	if err != nil {
		fmt.Fprintf(multiWriter, "%s 数据库连接失败: %s\n", symbols["[-]"]("[-]"), err)
		os.Exit(1)
	}

	var tableNames []map[string]interface{}
	db.Raw("SHOW TABLES").Scan(&tableNames)

	for _, v := range tableNames { // 循环从数据库取出的表map
		for _, s := range v { // 循环表map得到键值对
			sstring := s.(string) // 转换数据库名称为字符串
	
			if *table != "" && sstring != *table {
				continue
			}
	
			fmt.Fprintf(multiWriter, "%s 正在查询表： %s\n", symbols["[*]"]("[*]"), sstring)

			if *verbose {
				var data []map[string]interface{}
				db.Table(sstring).Find(&data)
			
				seenFields := make(map[string]bool)
				foundSensitiveInfo := false // 新增变量标记是否找到敏感信息
			
				for _, row := range data {
					for field, value := range row {
						for rulerName, rulerPattern := range ruler {
							var valueStr string
							switch v := value.(type) {
							case string:
								valueStr = v
							case int:
								valueStr = strconv.Itoa(v)
							case float64:
								valueStr = strconv.FormatFloat(v, 'f', -1, 64)
							case int32:
								valueStr = strconv.Itoa(int(v))
							case int64:
								valueStr = strconv.Itoa(int(v))
							}
			
							match, _ := regexp.MatchString(rulerPattern, valueStr)
							if match && !seenFields[field] {
								fmt.Fprintf(multiWriter, "%s 发现敏感信息: %s  表名称: %s  字段名称: %s  数据样例: %s\n",
									symbols["[!]"]("[!]"), rulerName, sstring, field, valueStr)	
								seenFields[field] = true // 标记此字段已输出
								foundSensitiveInfo = true // 找到敏感信息
							}
						}
					}
				}
			
				if len(data) > 0 {
					columns := make([]string, 0)
					for key := range data[0] {
						columns = append(columns, key)
					}
					if !foundSensitiveInfo {
						fmt.Fprintln(multiWriter, symbols["[-]"]("[-]"), "没有发现敏感信息,请手动确认")
					}
					printTableHeader(columns, multiWriter)
					for _, row := range data {
						printTableRow(row, columns, multiWriter)
					}
				}
			} else {
				var data []map[string]interface{}
				db.Table(sstring).Limit(1).Find(&data)
			
				seenFields := make(map[string]bool)
			
				if len(data) > 0 {
					for field, value := range data[0] {
						for rulerName, rulerPattern := range ruler {
							var valueStr string
							switch v := value.(type) {
							case string:
								valueStr = v
							case int:
								valueStr = strconv.Itoa(v)
							case float64:
								valueStr = strconv.FormatFloat(v, 'f', -1, 64)
							case int32:
								valueStr = strconv.Itoa(int(v))
							case int64:
								valueStr = strconv.Itoa(int(v))
							}
			
							match, _ := regexp.MatchString(rulerPattern, valueStr)
							if match && !seenFields[field] {
								fmt.Fprintf(multiWriter, "%s 发现敏感信息: %s  表名称: %s  字段名称: %s  数据样例: %s\n",
    								symbols["[!]"]("[!]"), rulerName, sstring, field, valueStr)
								seenFields[field] = true // 标记此字段已输出
							}
						}
					}
			
					columns := make([]string, 0)
					for key := range data[0] {
						columns = append(columns, key)
					}
					printTableHeader(columns, multiWriter)
			
					printTableRow(data[0], columns, multiWriter)
			
					fmt.Fprintln(multiWriter, "提示: 仅显示了第一行数据。使用 -v 参数可查看所有行。")
				} else {
					fmt.Fprintln(multiWriter, symbols["[-]"]("[-]"), "没有发现敏感信息")
				}
			}
		}
	}
	if *outputFile != "" {
		fmt.Fprintf(multiWriter, "%s 已导出报告文件: %s\n", symbols["[+]"]("[+]"), *outputFile)
	}
}
