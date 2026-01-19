package collections

import (
	"log"

	"github.com/StackExchange/wmi"
)

type ServiceData struct {
	Name        string `json:"service_name"`
	DisplayName string `json:"display_name"`
	Pid         int32  `json:"pid"`
	StartMode   string `json:"mode"`
	State       string `json:"state"`
	Status      string `json:"status"`
	Path        string `json:"path"`
}

type wmiService struct {
	Name        string
	DisplayName string
	ProcessId   int32
	StartMode   string
	State       string
	Status      string
	PathName    string
}

func SvcCollect() []ServiceData {
	var wmiResults []wmiService
	var services []ServiceData

	err := wmi.Query(
		"SELECT Name, DisplayName, ProcessId, StartMode, State, Status, PathName FROM Win32_Service",
		&wmiResults,
	)
	if err != nil {
		log.Printf("[-] Error retrieving services: %v", err)
		return services
	}

	for _, s := range wmiResults {
		services = append(services, ServiceData{
			Name:        s.Name,
			DisplayName: s.DisplayName,
			Pid:         s.ProcessId,
			StartMode:   s.StartMode,
			State:       s.State,
			Status:      s.Status,
			Path:        s.PathName,
		})
	}

	return services
}
