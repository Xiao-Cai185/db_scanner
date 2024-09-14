package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func displayLogo() {
	fmt.Println(`
 _____     ______     ______     ______     ______     __   __     __   __     ______     ______    
/\  __-.  /\  == \   /\  ___\   /\  ___\   /\  __ \   /\ "-.\ \   /\ "-.\ \   /\  ___\   /\  == \
\ \ \/\ \ \ \  __<   \ \___  \  \ \ \____  \ \  __ \  \ \ \-.  \  \ \ \-.  \  \ \  __\   \ \  __<
 \ \____-  \ \_____\  \/\_____\  \ \_____\  \ \_\ \_\  \ \_\\"\_\  \ \_\\"\_\  \ \_____\  \ \_\ \_\
  \/____/   \/_____/   \/_____/   \/_/\/_/   \/_/ \/_/  \/_/ \/_/   \/_/ \/_/   \/_____/   \/_/ /_/  
   Ver:2.1 Rebuild BY: Xiao_Cai185
  `)
}

func isHan(r rune) bool {
	return unicode.Is(unicode.Han, r)
}

func printTableHeader(columns []string) {
	fmt.Print("+")
	for range columns {
		fmt.Print("-------------------------+")
	}
	fmt.Println()
	fmt.Print("|")
	for _, col := range columns {
		fmt.Printf(" %-23s |", col)
	}
	fmt.Println()
	fmt.Print("+")
	for range columns {
		fmt.Print("-------------------------+")
	}
	fmt.Println()
}

func printTableRow(data map[string]interface{}, columns []string) {
	fmt.Print("|")
	for _, col := range columns {
		val, ok := data[col]
		if ok {
			switch v := val.(type) {
			case string:
				fmt.Printf(" %-23s |", v)
			case int, int32, int64:
				fmt.Printf(" %-23d |", v)
			case float64:
				fmt.Printf(" %-23.2f |", v)
			default:
				fmt.Printf(" %-23v |", v)
			}
		} else {
			fmt.Print(" %-23s |", "")
		}
	}
	fmt.Println()
	fmt.Print("+")
	for range columns {
		fmt.Print("-------------------------+")
	}
	fmt.Println()
}

func printDatabaseList(databases []string) {
	fmt.Print("+-------------------------+\n")
	fmt.Print("| Database Name           |\n")
	fmt.Print("+-------------------------+\n")
	for _, dbName := range databases {
		fmt.Printf("| %-23s |\n", dbName)
	}
	fmt.Print("+-------------------------+\n")
}

func main() {
	displayLogo()

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

	// 解析命令行参数
	flag.Parse()

	// 如果没有传入任何参数，输出帮助信息并退出
	if len(os.Args) == 1 || *help {
		fmt.Println("用法:")
		fmt.Println("  --host, -h         数据库主机地址")
		fmt.Println("  --user, -u         数据库用户名")
		fmt.Println("  --pass, -p         数据库密码")
		fmt.Println("  --port, -P         数据库端口号（默认：3306）")
		fmt.Println("  --char, -c         字符编码（默认：utf8mb4）")
		fmt.Println("  --database, --db   数据库名")
		fmt.Println("  --verbose, -v      是否显示符合条件的每一行表数据")
		fmt.Println("  --table, -t        指定要查询的表名")
		fmt.Println("  --showdbs, -s      展示所有数据库名称")
		fmt.Println("  --help             显示帮助信息")
		fmt.Println("例子:")
		fmt.Println("  1. 查询数据库列表:")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 -s")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password -s")
		fmt.Println("  2. 查询特定数据库:")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 --db testdb")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password --db testdb")
		fmt.Println("  3. 查询特定表:")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 --db testdb -t users")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password --db testdb -t users")
		fmt.Println("  4. 查询特定表并显示所有行数据:")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password -P 3306 -c utf8mb4 --db testdb -t users -v")
		fmt.Println("     db_scanner -h 127.0.0.1 -u root -p password --db testdb -t users -v")
		os.Exit(0)
	}

	// 如果指定了 --showdbs 参数，则仅展示数据库列表
	if *showdbs {
		if *host == "" || *user == "" || *pass == "" {
			fmt.Println("错误：必须提供数据库主机地址、用户名和密码。")
			fmt.Println("使用 --help 获取更多信息。")
			os.Exit(1)
		}

		mysqlpath := strings.Join([]string{*user, ":", *pass, "@tcp(", *host, ":", *port, ")/", "?charset=", *char}, "")

		db, err := gorm.Open(mysql.Open(mysqlpath), &gorm.Config{})
		if err != nil {
			fmt.Printf("数据库连接失败: %s\n", err)
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

		printDatabaseList(dbNames)
		os.Exit(0)
	}

	// 检查必须的参数是否缺失
	if *host == "" || *user == "" || *pass == "" || *database == "" {
		fmt.Println("错误：必须提供数据库主机地址、用户名、密码和数据库名。")
		fmt.Println("使用 --help 获取更多信息。")
		os.Exit(1)
	}

	// 加载规则文件
	viper.SetConfigName("ruler")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("无法读取规则文件: %s\n", err)
		os.Exit(1)
	}
	ruler := viper.GetStringMapString("ruler")

	// 拼接MySQL连接字符串
	mysqlpath := strings.Join([]string{*user, ":", *pass, "@tcp(", *host, ":", *port, ")/", *database, "?charset=", *char}, "")

	db, err := gorm.Open(mysql.Open(mysqlpath), &gorm.Config{}) // 链接数据库
	if err != nil {
		fmt.Printf("数据库连接失败: %s\n", err)
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

			fmt.Printf("正在查询表： %s\n", sstring)

			if *verbose {
				var data []map[string]interface{}
				db.Table(sstring).Find(&data)

				seenFields := make(map[string]bool)

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
								fmt.Printf("敏感信息: %s  表名称: %s  字段名称: %s  数据样例: %s\n",
									rulerName, sstring, field, valueStr)
								seenFields[field] = true // 标记此字段已输出
							}
						}
					}
				}

				if len(data) > 0 {
					columns := make([]string, 0)
					for key := range data[0] {
						columns = append(columns, key)
					}
					printTableHeader(columns)

					for _, row := range data {
						printTableRow(row, columns)
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
								fmt.Printf("敏感信息: %s  表名称: %s  字段名称: %s  数据样例: %s\n",
									rulerName, sstring, field, valueStr)
								seenFields[field] = true // 标记此字段已输出
							}
						}
					}

					columns := make([]string, 0)
					for key := range data[0] {
						columns = append(columns, key)
					}
					printTableHeader(columns)

					printTableRow(data[0], columns)

					fmt.Println("提示: 仅显示了第一行数据。使用 -v 参数可查看所有行。")
				}
			}
		}
	}
}
