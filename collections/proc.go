package collections

import (
	"github.com/shirou/gopsutil/v3/process"
)

type ProcessData struct {
	ProcessName string `json:"process_name"`
	Pid         int32  `json:"pid"`
	ParentName  string `json:"parent_name"`
	PPid        int32  `json:"ppid"`
	CmdLine     string `json:"cmd_line"`
	Username    string `json:"username"`
}

func ProcCollect() []ProcessData {
	var list []ProcessData

	procs, err := process.Processes()
	if err != nil {
		return list
	}

	for _, p := range procs {
		name, _ := p.Name()
		cmd, _ := p.Cmdline()
		username, _ := p.Username()
		ppid, _ := p.Ppid()

		parentName := "UNKNOWN"
		parentObj, err := p.Parent()
		if err == nil && parentObj != nil {
			pName, err := parentObj.Name()
			if err == nil {
				parentName = pName
			}
		}

		list = append(list, ProcessData{
			ProcessName: name,
			Pid:         p.Pid,
			ParentName:  parentName,
			PPid:        ppid,
			CmdLine:     cmd,
			Username:    username,
		})
	}
	return list
}
