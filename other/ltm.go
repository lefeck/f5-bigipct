package main

import (
	"flag"
	"fmt"
	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
	"github.com/xuri/excelize/v2"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
 tips : 这是第二种实现方式，这种方式是通过读取Excel中Slice-->Map-->Struct中，来实现数据批量导入操作。
*/

//const HashedPassword = "$2a$10$tnkFMyd/VOWbN5JBWzt4oO5Hc0S6RryXH8KtJsktsUArAPbwZ6dLy"

var (
	host     string
	user     string
	password string
	file     string
	sheet    string
	vss      ltm.VirtualServer
)

type VirtualServer struct {
	Virtual_Name      string
	Vs_Destination    string
	Vs_IP_Protocol    string
	Profiles          string
	Translate_Address string
	Translate_Port    string
	Snat_Type         string
	Persistence       string
	Pool_Name         string
	Pool_Member       string
	Pool_Monitor      string
	Pool_Lbmode       string
}

type ExcelData interface {
	// 把excel中每行数据转换成map
	CreateMap(arr []string) map[string]interface{}
	ChangeTime(source string) time.Time
}

type ExcelStrcut struct {
	// 二维数组
	temp  [][]string
	Model interface{}
	Info  []map[string]string
}

func init() {
	flag.StringVar(&host, "a", "192.168.1.1", "the host ip address.")
	flag.StringVar(&user, "u", "admin", "specifies the username of login host.")
	flag.StringVar(&password, "p", "admin", "specifies the password of login host.")
	flag.StringVar(&file, "f", "/tmp/test.xlsx", "specifies an alternative configuration file.")
	flag.StringVar(&sheet, "s", "Sheet1", "specifies the table name of the workbook.")

	flag.Parse()

	// default password : "7923w4T28M"
	//password, _ := gopass.GetPasswdPrompt("please input password: ", true, os.Stdin, os.Stdout)
	//if err := bcrypt.CompareHashAndPassword([]byte(HashedPassword), []byte(password)); err != nil {
	//	log.Fatalf("login password err: %v ", err)
	//}
}

func NewF5Client() (*f5.Client, error) {
	hosts := fmt.Sprintf("https://" + host)
	client, err := f5.NewBasicClient(hosts, user, password)
	//client, err := f5.NewBasicClient("https://192.168.10.84", "admin", "admin")
	client.DisableCertCheck()
	if err != nil {
		fmt.Println(err)
	}
	return client, nil
}

// 读取Excel中数据 转换成二维数组
func (excel *ExcelStrcut) ReadExcel(file string) *ExcelStrcut {
	f, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(rows)
	excel.temp = rows
	return excel
}

//将二维数组中的每行转成对应的map
func (excel *ExcelStrcut) CreateMap() *ExcelStrcut {
	//利用反射得到字段名
	for _, v := range excel.temp {
		//将二维数组的每行转成对应切片类型的map
		var info = make(map[string]string)
		for i := 0; i < reflect.ValueOf(excel.Model).NumField(); i++ {
			obj := reflect.TypeOf(excel.Model).Field(i)
			//fmt.Printf("key:%s--val:%s\n", obj.Name, v[i])
			info[obj.Name] = v[i]
		}
		excel.Info = append(excel.Info, info)
	}
	return excel
}

// 时间做格式化
func (excel *ExcelStrcut) ChangeTime(source string) time.Time {
	times, err := time.Parse("2006-01-02", source)
	if err != nil {
		log.Fatalf("转换时间错误:%s", err)
	}
	return times
}

func (vs *VirtualServer) Create(client *f5.Client) (err error) {
	tx, err := client.Begin()
	if err != nil {
		log.Fatal(err)
	}

	ltmclient := ltm.New(tx)

	members := StringToSlice(vs.Pool_Member)

	pool := ltm.Pool{
		Name:              vs.Pool_Name,
		Monitor:           vs.Pool_Monitor,
		LoadBalancingMode: vs.Pool_Lbmode,
		Members:           members,
	}

	if err := ltmclient.Pool().Create(pool); err != nil {
		log.Fatal(err)
	}

	profile := StringToSlice(vs.Profiles)

	if vs.Persistence == "none" {
		vss = ltm.VirtualServer{
			Name:             vs.Virtual_Name,
			Destination:      vs.Vs_Destination,
			IPProtocol:       vs.Vs_IP_Protocol,
			Profiles:         profile,
			TranslateAddress: vs.Translate_Address,
			TranslatePort:    vs.Translate_Port,
			//Persistences:             []ltm.Persistence{{Name: vs.Persistence}},
			Pool:                     vs.Pool_Name,
			SourceAddressTranslation: ltm.SourceAddressTranslation{Type: vs.Snat_Type, Pool: ""},
		}
	} else {
		vss = ltm.VirtualServer{
			Name:                     vs.Virtual_Name,
			Destination:              vs.Vs_Destination,
			IPProtocol:               vs.Vs_IP_Protocol,
			Profiles:                 profile,
			TranslateAddress:         vs.Translate_Address,
			TranslatePort:            vs.Translate_Port,
			Persistences:             []ltm.Persistence{{Name: vs.Persistence}},
			Pool:                     vs.Pool_Name,
			SourceAddressTranslation: ltm.SourceAddressTranslation{Type: vs.Snat_Type, Pool: ""},
		}
	}

	if err := ltmclient.Virtual().Create(vss); err != nil {
		log.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("virtualserver name %s create success.\n", vs.Virtual_Name)

	return nil
}

func (excel *ExcelStrcut) MapToStruct(vs *VirtualServer, client *f5.Client) *ExcelStrcut {
	//忽略标题行
	for i := 1; i < len(excel.Info); i++ {
		t := reflect.ValueOf(vs).Elem()
		for k, v := range excel.Info[i] {
			// 从map中读取出字段的值和类型
			//fmt.Println("key:%v---val:%v", t.FieldByName(k), t.FieldByName(k).Kind())
			switch t.FieldByName(k).Kind() {
			case reflect.String:
				//把map中的value写入到struct
				t.FieldByName(k).Set(reflect.ValueOf(v))
			case reflect.Float64:
				strToFloat64, err := strconv.ParseFloat(v, 64)
				if err != nil {
					log.Printf("string to float64 err：%v", err)
				}
				t.FieldByName(k).Set(reflect.ValueOf(strToFloat64))
			case reflect.Uint64:
				strToUint64, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					log.Printf("string to uint64 err：%v", err)
				}
				t.FieldByName(k).Set(reflect.ValueOf(strToUint64))
			case reflect.Struct:
				times, err := time.Parse("2006-01-02", v)
				if err != nil {
					log.Printf("string to time err：%v", err)
				}
				t.FieldByName(k).Set(reflect.ValueOf(times))
			default:
				fmt.Println("type err")
			}
		}
		if err := vs.Create(client); err != nil {
			log.Fatal(err)
		}
	}
	return excel
}

func StringToSlice(src string) []string {
	var result []string
	str := strings.Replace(src, "\n", " ", 1)
	s := DeleteExtraSpace(str)
	splitSlice := strings.Split(s, " ")
	return append(result, splitSlice...)
}

func DeleteExtraSpace(s string) string {
	s1 := strings.Replace(s, "	", " ", -1)
	regstr := "\\s{2,}"
	reg, _ := regexp.Compile(regstr)
	s2 := make([]byte, len(s1))
	copy(s2, s1)
	spc_index := reg.FindStringIndex(string(s2))
	for len(spc_index) > 0 {
		s2 = append(s2[:spc_index[0]+1], s2[spc_index[1]:]...)
		spc_index = reg.FindStringIndex(string(s2))
	}
	return string(s2)
}

func main() {
	client, _ := NewF5Client()
	excel := ExcelStrcut{}
	vs := VirtualServer{}
	excel.Model = vs
	excel.ReadExcel(file).CreateMap().MapToStruct(&vs, client)
}
