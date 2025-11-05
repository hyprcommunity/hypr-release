package backend

import (
	"fmt"
	"time"

	"hyprrelease/api/releases/check/hyprland"
	"hyprrelease/api/releases/updateing"
	"hyprrelease/summaryofversion"
)

// UpdateDotfileAndSystem : dotfile + hyprland sistemini karşılaştırır ve gerekirse günceller
func UpdateDotfileAndSystem(dotfileName string) error {
	fmt.Printf("[hyprrelease-update] checking for updates: %s\n", dotfileName)

	// Dotfile registry'den çekiliyor
	var selected *summaryofversion.Dotfile
	for _, d := range summaryofversion.Registry {
		if d.Name == dotfileName {
			selected = &d
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("dotfile not found: %s", dotfileName)
	}

	// Hyprland sistemini kontrol et
	fmt.Println("[hyprrelease-update] scanning system components...")
	components, logText, err := hyprland.CheckHyprSystem()
	if err != nil {
		fmt.Println("⚠️ failed to check Hypr system:", err)
	} else {
		fmt.Println(logText)
	}

	// Güncelleme olup olmadığını analiz et
	updatesAvailable := false
	for _, c := range components {
		if c.UpdateAvailable {
			updatesAvailable = true
			break
		}
	}

	// Güncelleme varsa veya kullanıcı isterse Dotfile yeniden kurulabilir
	if updatesAvailable {
		fmt.Println("[hyprrelease-update] system updates detected.")
	} else {
		fmt.Println("[hyprrelease-update] all system components up to date.")
	}

	// Dotfile metadata’yı oluştur (güncelleme sırasında sürüm bilgisi yazar)
	versionMain := "unknown"
	versionBuild := time.Now().Format("20060102")
	branch := selected.Branch
	releaseChannel := "stable"
	commitsBehind := "0"

	err = updateing.WriteMetaFile(
		selected.Name,
		versionMain,
		versionBuild,
		branch,
		releaseChannel,
		commitsBehind,
	)
	if err != nil {
		fmt.Println("⚠️ failed to update hyprland-release metadata:", err)
	} else {
		fmt.Println("[hyprrelease-update] /etc/hyprland-release updated")
	}

	// Güncelleme işlemi (kullanıcı onayı alınabilir)
	fmt.Print("Do you want to reinstall or update this dotfile? [y/N]: ")
	var resp string
	fmt.Scanln(&resp)
	if resp != "y" && resp != "Y" {
		fmt.Println("[hyprrelease-update] skipped reinstall")
		return nil
	}

	fmt.Println("[hyprrelease-update] reinstalling dotfile from registry...")
	if err := InstallFromRegistry(dotfileName); err != nil {
		return fmt.Errorf("installation failed: %v", err)
	}

	fmt.Println("[hyprrelease-update] done.")
	return nil
}
