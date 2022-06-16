package ltm

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
	"github.com/xuri/excelize/v2"
)

/*
 tips : 这是第一种实现方式，这种方式是通过读取Excel中Slice-->Struct中，来实现数据批量导入操作。
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

func NewVirtualServers() *VirtualServer {
	return &VirtualServer{}
}

func (vs *VirtualServer) Exec(client *f5.Client) (err error) {
	//files, err := excelize.OpenFile("./create.xlsx")
	files, err := excelize.OpenFile(file)
	if err != nil {
		fmt.Println(err)
	}
	defer files.Close()

	rows, err := files.GetRows(sheet)
	if err != nil {
		fmt.Println(err)
	}

	for key, row := range rows {
		if key > 0 {
			if err := SliceToStruct(row, vs); err != nil {
				log.Fatal(err)
			}
			if err := vs.Create(client); err != nil {
				log.Fatal(err)
			}
		}
	}
	return nil
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

func SliceToStruct(arr []string, u interface{}) error {
	valueOf := reflect.ValueOf(u)
	if valueOf.Kind() != reflect.Ptr {
		return errors.New("must ptr")
	}
	valueOf = valueOf.Elem()
	if valueOf.Kind() != reflect.Struct {
		return errors.New("must struct")
	}
	for i := 0; i < valueOf.NumField(); i++ {
		if i >= len(arr) {
			break
		}
		val := arr[i]
		if val != "" && reflect.ValueOf(val).Kind() == valueOf.Field(i).Kind() {
			valueOf.Field(i).Set(reflect.ValueOf(val))
		}
	}
	return nil
}

func StringToSlice(src string) []string {
	var result []string
	str := strings.Replace(src, "\n", " ", -1)
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
