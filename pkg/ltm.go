package pkg

import (
	"errors"
	//"f5ltm"
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

var (
	vss ltm.VirtualServer
)

type VirtualServers struct {
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

func NewVirtualServers() *VirtualServers {
	return &VirtualServers{}
}

func (vs *VirtualServers) Import(client *f5.Client) (err error) {
	//files, err := excelize.OpenFile("./create.xlsx")
	files, err := excelize.OpenFile(File)
	if err != nil {
		log.Fatalf("open file failed: %s", err)
	}
	defer files.Close()

	rows, err := files.GetRows(Sheet)
	if err != nil {
		log.Fatalf("the sheet is not exist: %s", err)
	}

	for key, row := range rows {
		if key > 0 {
			if err = SliceToStruct(row, vs); err != nil {
				log.Fatalf("configure parsing failed", err)
			}
			if err = vs.Create(client); err != nil {
				log.Fatalf("create configure failed", err)
			}
		}
	}
	return nil
}

func (vs *VirtualServers) Create(client *f5.Client) (err error) {
	tx, err := client.Begin()
	if err != nil {
		log.Fatalf("clients open transaction: %s", err)
	}
	ltmclient := ltm.New(tx)
	members := StringToSlice(vs.Pool_Member)
	pool := ltm.Pool{
		Name:              vs.Pool_Name,
		Monitor:           vs.Pool_Monitor,
		LoadBalancingMode: vs.Pool_Lbmode,
		Members:           members,
	}
	if err = ltmclient.Pool().Create(pool); err != nil {
		log.Fatalf("clients create  pool failed: %s", err)
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
			//Persistences:             []pkg.Persistence{{Name: vs.Persistence}},
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
	if err = ltmclient.Virtual().Create(vss); err != nil {
		log.Fatalf("clients create virtualserver failed: %s", err)
	}
	if err = tx.Commit(); err != nil {
		log.Fatalf("clients commits transaction: %s", err)
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
