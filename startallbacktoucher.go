// THANKS ALOT TO Aetherinox for making this possible and opensourcing his version which i simply "Translated" to go ily so much bro feel free to check his [github](https://github.com/Aetherinox/utility-startallback/tree/main)
// This is an extremely minimal version of the original script just barebones cui and no bloatware.
package main

import (
	"golang.org/x/sys/windows/registry"
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	err := start("auto")
	if err != nil {
		fmt.Println("error", err)
	}
}

func start(sel string) error {
	// sets some paths and values 
	lad := os.Getenv("LOCALAPPDATA")
	pf := os.Getenv("ProgramFiles")
	pf86 := os.Getenv("ProgramFiles(x86)")
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	dll := "StartAllBackX64.dll"
	dlls := []string{
		filepath.Join(lad, "StartAllBack", dll),
		filepath.Join(pf, "StartAllBack", dll),
		filepath.Join(pf86, "StartAllBack", dll),
		filepath.Join(dir, "StartAllBack", dll),
	}
	var found []string
	for _, p := range dlls {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}
	if sel != "auto" {
		if _, err := os.Stat(sel); err == nil {
			found = []string{sel}
		}
	}
	if len(found) == 0 {
		return fmt.Errorf("No valid DLL paths found")
	}
	// disables auto restart
	err := setReg(`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`, "AutoRestartShell", 0)
	if err != nil {
		return err
	}

	// Kills explorer and startallbackcfg
	exec.Command("taskkill", "/f", "/im", "explorer.exe").Run()
	exec.Command("taskkill", "/f", "/im", "StartAllBackCfg.exe").Run()

	// Patches the DLLS
	for _, dllPath := range found {
		err := patchDLL(dllPath)
		if err != nil {
			fmt.Printf("Failed to patch %s: %v\n", dllPath, err)
			fmt.Printf("RETRYING IN 5 SECONDS\n")
			time.Sleep(5 * time.Second)
			patchDLL(dllPath)
		}
	}
	// Starts everything back up (and ofc enables back auto restart)
	exec.Command("explorer.exe").Start()
	err = setReg(`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Winlogon`, "AutoRestartShell", 1)
	if err != nil {
		return err
	}
	fmt.Println("Patch complete successfully injected attistartallback (im just playing it means that your startallback has been patched)")
	return nil
}

func patchDLL(dllPath string) error {
	// restores backup if exists
	bak := dllPath + ".bak"
	if _, err := os.Stat(bak); err == nil {
		os.Remove(dllPath)
		os.Rename(bak, dllPath)
	}

	// reads and backups original dll
	data, err := ioutil.ReadFile(dllPath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(bak, data, 0644)
	if err != nil {
		return err
	}
	// defines and converts hex into sort of slices (which it replaces the pattern and finishes writing the patched dll)
	orig := "48895C2408555657488DAC2470FFFFFF"
	patch := "67C70101000000B801000000C3909090"
	obytes, _ := hex.DecodeString(orig)
	pbytes, _ := hex.DecodeString(patch)
	index := bytes.Index(data, obytes)
	if index == -1 {
		return fmt.Errorf("Pattern not found in %s", dllPath)
	}
	copy(data[index:], pbytes)
	err = ioutil.WriteFile(dllPath, data, 0644)
	if err != nil {
		return err
	}

	// Launches the exe of startallback back up
	sab := filepath.Join(filepath.Dir(dllPath), "StartAllBack.exe")
	if _, err := os.Stat(sab); err == nil {
		exec.Command(sab).Start()
	}

	return nil
}

func setReg(path, name string, value uint32) error {
	key, exists, err := registry.CreateKey(registry.LOCAL_MACHINE, path, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %v", err)
	}
	defer key.Close()

	if !exists {
		return fmt.Errorf("registry path does not exist: %s", path)
	}

	if err := key.SetDWordValue(name, value); err != nil {
		return fmt.Errorf("failed to set registry value: %v", err)
	}

	return nil
}
