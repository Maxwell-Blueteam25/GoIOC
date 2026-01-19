package collections

import (
	"golang.org/x/sys/windows/registry"
)

type RegData struct {
	Hive  string `json:"hive"`
	Path  string `json:"key_path"`
	Name  string `json:"value_name"`
	Value string `json:"value_data"`
}

type targetKey struct {
	Hive registry.Key
	Name string // String representation of Hive
	Path string
}

func RegCollect() []RegData {
	var results []RegData

	targets := []targetKey{
		{registry.LOCAL_MACHINE, "HKLM", `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`},
		{registry.LOCAL_MACHINE, "HKLM", `SOFTWARE\Microsoft\Windows\CurrentVersion\RunOnce`},
		{registry.LOCAL_MACHINE, "HKLM", `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`}, // Shell/Userinit
		{registry.CURRENT_USER, "HKCU", `Software\Microsoft\Windows\CurrentVersion\Run`},
		{registry.CURRENT_USER, "HKCU", `Software\Microsoft\Windows\CurrentVersion\RunOnce`},
	}

	for _, t := range targets {
		k, err := registry.OpenKey(t.Hive, t.Path, registry.READ)
		if err != nil {
			continue
		}

		names, err := k.ReadValueNames(-1)
		if err != nil {
			k.Close()
			continue
		}

		for _, name := range names {
			val, _, err := k.GetStringValue(name)
			if err != nil {
				continue
			}

			results = append(results, RegData{
				Hive:  t.Name,
				Path:  t.Path,
				Name:  name,
				Value: val,
			})
		}
		k.Close()
	}

	return results
}
