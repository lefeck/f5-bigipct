package pkg

import (
	"f5ltm/conf"
	"fmt"
	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
	"github.com/xuri/excelize/v2"
	"log"
	"net/url"
	"reflect"
	"strings"
)

var (
	files *excelize.File
)

func NewVirtualServer() *VirtualServer {
	return &VirtualServer{}
}

func NewF5Client() (*f5.Client, error) {
	hosts := fmt.Sprintf("https://" + conf.Host)
	client, err := f5.NewBasicClient(hosts, conf.Username, conf.Password)
	//clients, err := f5.NewBasicClient("https://192.168.10.84", "admin", "admin")
	client.DisableCertCheck()
	if err != nil {
		log.Fatalf("clients connect to f5 device failed: %s", err)
	}
	return client, nil
}

type VirtualServer struct {
	Partition         string `json:"partition" xlsx:"partition"`
	VirtualServerName string `json:"virtualservername" xlsx:"virtualservername"`
	Destination       string `json:"destination" xlsx:"destination"`
	Description       string `json:"description" xlsx:"description"`
	Source            string `json:"source" xlsx:"source"`
	VS_IP_Protocol    string `json:"vs_ip_protocol" xlsx:"vs_ip_protocol"`
	Profiles          string `json:"profiles" xlsx:"profiles"`
	PoolName          string `json:"pool_name" xlsx:"pool_name"`
	PoolMembers       string `json:"pool_members" xlsx:"pool_members"`
	SnatType          string `json:"snat_type" xlsx:"snat_type"`
	SnatPool          string `json:"snatpool" xlsx:"snat_pool"`
	IRules            string `json:"irules" xlsx:"irules"`
	Monitors          string `json:"monitors" xlsx:"monitors"`
}

func (vs VirtualServer) Write(client *f5.Client) (err error) {
	tx, err := client.Begin()
	if err != nil {
		log.Fatalf("clients open transaction: %s", err)
	}
	ltmclient := ltm.New(tx)
	if err := WriteVirtualServerToXlsx(vs, conf.File, ltmclient); err != nil {
		log.Fatalf("write virtualserver to xlsx failed: %s", err)
	}
	if err := WriteProfiesToXlsx(conf.File, ltmclient); err != nil {
		log.Fatalf("write profiles to xlsx failed: %s", err)
	}
	if err := WriteMonitorsToXlsx(conf.File, ltmclient); err != nil {
		log.Fatalf("write monitors to xlsx failed: %s", err)
	}
	return nil
}

func WriteVirtualServerToXlsx(vs VirtualServer, file string, ltmclient ltm.LTM) error {
	var result []interface{}
	vslist, _ := ltmclient.Virtual().ListAll()
	for _, value := range vslist.Items {
		pool := StringSplitSubString(value.Pool)
		pools, _ := ltmclient.PoolMembers().ListAll(pool)
		var poolmembernames []string
		for _, poolmember := range pools.Items {
			poolmembernames = append(poolmembernames, poolmember.Name)
		}
		snatpoolname := StringSplitSubString(value.SourceAddressTranslation.Pool)
		vs = StructToStruct(value, snatpoolname, poolmembernames)
		result = append(result, vs)
	}
	files = WriteXlsx(conf.Sheet, result)
	if err := files.SaveAs(file); err != nil {
		log.Fatalf("file save failed: %s", err)
	}
	return nil
}

func WriteProfiesToXlsx(file string, ltmclient ltm.LTM) error {
	files, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer files.Close()

	params := url.Values{}
	params.Set("expandSubcollections", "true")

	vslists, _ := ltmclient.Virtual().ListAllWithParams(params)

	for key, value := range vslists.Items {
		profiles := value.ProfilesReference.Profiles
		var proFiles []string
		for _, profile := range profiles {
			proFiles = append(proFiles, profile.Name)
		}
		CreateExcelSlice(files, proFiles, key)
	}
	if err := files.SaveAs(file); err != nil {
		log.Fatalf("file save failed: %s", err)
	}
	return nil
}

func WriteMonitorsToXlsx(file string, ltmclient ltm.LTM) error {
	files, err := excelize.OpenFile(file)
	if err != nil {
		log.Fatalln(err)
	}
	defer files.Close()

	poollist, _ := ltmclient.Pool().ListAll()
	for key, v := range poollist.Items {
		CreateExcelString(files, v.Monitor, key)
	}

	if err := files.SaveAs(file); err != nil {
		log.Fatalf("file save failed: %s", err)
	}
	return nil
}

func CreateExcelString(f *excelize.File, src string, i int) error {
	str := StringSplitSubString(src)
	if err := f.SetCellValue(conf.Sheet, fmt.Sprintf("%s%d", "M", i+2), str); err != nil {
		return err
	}
	return nil
}

func CreateExcelSlice(f *excelize.File, src []string, i int) error {
	str := SliceToString(src)
	if err := f.SetCellValue(conf.Sheet, fmt.Sprintf("%s%d", "G", i+2), str); err != nil {
		return err
	}
	return nil
}

func WriteXlsx(sheet string, records interface{}) *excelize.File {
	xlsx := excelize.NewFile()
	index := xlsx.NewSheet(sheet)
	xlsx.SetActiveSheet(index)
	firstCharacter := 65
	t := reflect.TypeOf(records)

	if t.Kind() != reflect.Slice {
		return xlsx
	}

	s := reflect.ValueOf(records)

	for i := 0; i < s.Len(); i++ {
		elem := s.Index(i).Interface()
		elemType := reflect.TypeOf(elem)
		elemValue := reflect.ValueOf(elem)
		for j := 0; j < elemType.NumField(); j++ {
			field := elemType.Field(j)
			tag := field.Tag.Get("xlsx")
			name := tag
			column := string(firstCharacter + j)
			if tag == "" {
				continue
			}
			// 设置表头
			if i == 0 {
				xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+1), name)
			}
			// 设置内容
			xlsx.SetCellValue(sheet, fmt.Sprintf("%s%d", column, i+2), elemValue.Field(j).Interface())
		}
	}
	return xlsx
}

func StringSplitSubString(src string) (des string) {
	str := strings.SplitN(src, "/", -1)
	return str[len(str)-1]
}

func SliceToString(src []string) string {
	return strings.Join(src, " ")
}

func StructToStruct(server ltm.VirtualServer, snatpoolname string, poolmembers []string) VirtualServer {
	poolName := StringSplitSubString(server.Pool)
	destination := StringSplitSubString(server.Destination)
	poolMembers := SliceToString(poolmembers)
	irulesutils := SliceToString(server.Rules)
	irules := StringSplitSubString(irulesutils)

	return VirtualServer{
		Partition:         server.Partition,
		VirtualServerName: server.Name,
		Destination:       destination,
		Description:       server.Description,
		Source:            server.Source,
		VS_IP_Protocol:    server.IPProtocol,
		PoolName:          poolName,
		PoolMembers:       poolMembers,
		SnatPool:          snatpoolname,
		SnatType:          server.SourceAddressTranslation.Type,
		IRules:            irules,
	}
}
