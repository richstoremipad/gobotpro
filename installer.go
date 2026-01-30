package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

// --- KONFIGURASI ---
const EXTENSION_ID = "bncdgnpdlnmlilbdkkejihnjigdbfiaj" // ID BARU
const HOST_NAME = "com.automa.filefetcher"

//go:embed assets/*
var embeddedFiles embed.FS

// Global Vars
var (
	myWindow      fyne.Window
	mainContainer *fyne.Container
)

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme()) // Wajib Dark Mode biar pro

	myWindow = myApp.NewWindow("GoBot Driver Installer")
	myWindow.Resize(fyne.NewSize(600, 450))
	myWindow.SetFixedSize(true) // Kunci ukuran window agar layout tidak berantakan
	myWindow.CenterOnScreen()

	showWelcomeScreen()
	myWindow.ShowAndRun()
}

// ==========================================
// 🎨 1. WELCOME SCREEN (PROFESIONAL)
// ==========================================
func showWelcomeScreen() {
	// --- HEADER SECTION ---
	lblTitle := canvas.NewText("GoBot Driver Setup", color.White)
	lblTitle.TextSize = 28
	lblTitle.TextStyle = fyne.TextStyle{Bold: true}
	lblTitle.Alignment = fyne.TextAlignCenter

	lblSub := canvas.NewText("Native Automation Host Installer", color.NRGBA{R: 160, G: 160, B: 160, A: 255})
	lblSub.TextSize = 14
	lblSub.Alignment = fyne.TextAlignCenter

	// Info OS dengan kotak background tipis (Card effect simulasi)
	lblOS := widget.NewLabel(fmt.Sprintf("System: %s (%s)", runtime.GOOS, runtime.GOARCH))
	lblOS.Alignment = fyne.TextAlignCenter
	
	headerContainer := container.NewVBox(
		layout.NewSpacer(),
		lblTitle,
		lblSub,
		widget.NewIcon(theme.ComputerIcon()),
		lblOS,
		layout.NewSpacer(),
	)

	// --- BUTTON SECTION (FIXED SIZE) ---
	// Tombol Utama
	btnInstall := widget.NewButton("INSTALL DRIVER", func() {
		showLoadingScreen()
		go processInstallation("")
	})
	btnInstall.Icon = theme.ConfirmIcon()
	btnInstall.Importance = widget.HighImportance // Warna Biru Solid

	// Tombol Manual
	btnManual := widget.NewButton("Manual Browse (Portable)", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				showLoadingScreen()
				go processInstallation(uri.Path())
			}
		}, myWindow)
	})
	btnManual.Icon = theme.FolderOpenIcon()
	btnManual.Importance = widget.LowImportance // Warna Abu/Outline

	// LOGIKA AGAR TOMBOL TIDAK LEBAR
	// Kita gunakan GridWrap dengan ukuran tetap (misal: lebar 280, tinggi 45)
	var buttonGroup *fyne.Container
	
	if runtime.GOOS == "windows" {
		buttonGroup = container.NewGridWrap(
			fyne.NewSize(280, 45), // <-- INI KUNCINYA AGAR TOMBOL RAPI
			btnInstall,
			btnManual,
		)
	} else {
		buttonGroup = container.NewGridWrap(
			fyne.NewSize(280, 45),
			btnInstall,
		)
	}

	// Bungkus buttonGroup di tengah layar
	centerButtons := container.NewCenter(buttonGroup)

	// --- FOOTER SECTION (DISCLAIMER) ---
	lblCredit := canvas.NewText("Code by Rich Dev", color.NRGBA{R: 100, G: 100, B: 100, A: 255})
	lblCredit.TextSize = 11
	lblCredit.Alignment = fyne.TextAlignCenter
	lblCredit.TextStyle = fyne.TextStyle{Monospace: true}
	
	footerContainer := container.NewVBox(
		widget.NewSeparator(),
		container.NewPadded(lblCredit),
	)

	// --- FINAL LAYOUT (Border Layout) ---
	// Top: Header, Bottom: Footer, Center: Tombol
	content := container.NewBorder(
		headerContainer, // Atas
		footerContainer, // Bawah
		nil, nil,        // Kiri Kanan
		centerButtons,   // Tengah (Tombol)
	)

	// Tambahkan Padding Luar agar tidak mepet pinggir window
	myWindow.SetContent(container.NewPadded(content))
}

// ==========================================
// 🔄 2. LOADING SCREEN (CLEAN)
// ==========================================
func showLoadingScreen() {
	// Spinner besar
	spinner := widget.NewActivity()
	spinner.Start()

	// Label Status
	lblStatus := widget.NewLabel("Memulai proses instalasi...")
	lblStatus.Alignment = fyne.TextAlignCenter

	// Credit Footer tetap ada
	lblCredit := canvas.NewText("Code by Rich Dev", color.NRGBA{R: 80, G: 80, B: 80, A: 255})
	lblCredit.TextSize = 10
	lblCredit.Alignment = fyne.TextAlignCenter

	// Layout Tengah
	centerContent := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(spinner), // Spinner di tengah
		layout.NewSpacer(),
		lblStatus,
		layout.NewSpacer(),
	)

	content := container.NewBorder(
		nil,
		container.NewPadded(lblCredit),
		nil, nil,
		centerContent,
	)

	myWindow.SetContent(container.NewPadded(content))
	
	// Simpan referensi untuk update status
	mainContainer = centerContent
}

func updateStatus(msg string) {
	// Update label teks (child ke-3 dari VBox di atas)
	if mainContainer != nil && len(mainContainer.Objects) >= 3 {
		if lbl, ok := mainContainer.Objects[3].(*widget.Label); ok {
			lbl.SetText(msg)
		}
	}
}

// ==========================================
// ✅ 3. SUCCESS SCREEN (RAPI)
// ==========================================
func showSuccessScreen(logs string) {
	// Icon Sukses Besar
	iconCheck := widget.NewIcon(theme.ConfirmIcon())
	
	lblTitle := canvas.NewText("INSTALASI SELESAI", color.NRGBA{R: 76, G: 175, B: 80, A: 255}) // Warna Hijau
	lblTitle.TextSize = 22
	lblTitle.TextStyle = fyne.TextStyle{Bold: true}
	lblTitle.Alignment = fyne.TextAlignCenter

	// Kotak Log (Rapi dengan border)
	entryLog := widget.NewMultiLineEntry()
	entryLog.SetText(logs)
	entryLog.Disable()
	entryLog.TextStyle = fyne.TextStyle{Monospace: true}
	
	// Tombol Keluar (Fixed Size)
	btnClose := widget.NewButton("TUTUP APLIKASI", func() {
		myWindow.Close()
	})
	btnClose.Importance = widget.HighImportance
	
	btnWrapper := container.NewCenter(container.NewGridWrap(
		fyne.NewSize(200, 40),
		btnClose,
	))

	// Layout
	topPart := container.NewVBox(
		container.NewCenter(iconCheck),
		lblTitle,
		widget.NewSeparator(),
	)

	content := container.NewBorder(
		topPart,      // Atas
		btnWrapper,   // Bawah (Tombol)
		nil, nil,
		container.NewPadded(container.NewScroll(entryLog)), // Tengah (Log)
	)

	myWindow.SetContent(container.NewPadded(content))
}

// ==========================================
// ⚙️ LOGIKA SYSTEM (BACKEND)
// ==========================================
func processInstallation(customPath string) {
	time.Sleep(800 * time.Millisecond) // UX Delay
	
	updateStatus("Menganalisis sistem...")
	time.Sleep(800 * time.Millisecond)

	updateStatus("Mengekstrak file driver...")
	installDir, err := getPermanentInstallDir()
	if err != nil {
		dialog.ShowError(err, myWindow); return
	}
	os.MkdirAll(installDir, 0755)

	fileName := "gobot"
	if runtime.GOOS == "windows" { fileName = "gobot.exe" }

	fileData, err := fs.ReadFile(embeddedFiles, "assets/"+fileName)
	if err != nil {
		updateStatus("Error: Aset 'assets/" + fileName + "' tidak ditemukan!")
		time.Sleep(3 * time.Second)
		myWindow.Close()
		return
	}

	finalBinPath := filepath.Join(installDir, fileName)
	ioutil.WriteFile(finalBinPath, fileData, 0755)

	updateStatus("Mendaftarkan Manifest...")
	time.Sleep(500 * time.Millisecond)

	var logs string
	if runtime.GOOS == "windows" {
		logs = installWindowsLogic(finalBinPath, customPath)
	} else {
		logs = installLinuxLogic(finalBinPath)
	}

	updateStatus("Menyelesaikan proses...")
	time.Sleep(1 * time.Second)

	showSuccessScreen(logs)
}

func getPermanentInstallDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil { return "", err }
	return filepath.Join(configDir, "GoBotAutomation"), nil
}

func createManifest(savePath, binaryPath string) error {
	data := map[string]interface{}{
		"name":            HOST_NAME,
		"description":     "GoBot Driver",
		"path":            binaryPath,
		"type":            "stdio",
		"allowed_origins": []string{"chrome-extension://" + EXTENSION_ID + "/"},
	}
	file, _ := json.MarshalIndent(data, "", "  ")
	return ioutil.WriteFile(savePath, file, 0644)
}

func installLinuxLogic(binPath string) string {
	home, _ := os.UserHomeDir()
	targets := []string{
		filepath.Join(home, ".config/google-chrome/NativeMessagingHosts"),
		filepath.Join(home, ".config/chromium/NativeMessagingHosts"),
		filepath.Join(home, ".config/BraveSoftware/Brave-Browser/NativeMessagingHosts"),
	}
	log := ""
	count := 0
	for _, dir := range targets {
		if _, err := os.Stat(filepath.Dir(dir)); os.IsNotExist(err) { continue }
		os.MkdirAll(dir, 0755)
		manifestPath := filepath.Join(dir, HOST_NAME+".json")
		if createManifest(manifestPath, binPath) == nil {
			log += fmt.Sprintf("✅ Inject Success: %s\n", dir)
			count++
		}
	}
	if count == 0 { log += "⚠️ Tidak ditemukan instalasi browser standar.\n" }
	return log
}

func installWindowsLogic(binPath, customPath string) string {
	log := ""
	if customPath == "" {
		manifestPath := filepath.Join(filepath.Dir(binPath), "manifest.json")
		createManifest(manifestPath, binPath)
		keys := []string{
			`HKCU\Software\Google\Chrome\NativeMessagingHosts\` + HOST_NAME,
			`HKCU\Software\Microsoft\Edge\NativeMessagingHosts\` + HOST_NAME,
		}
		for _, k := range keys {
			exec.Command("reg", "add", k, "/ve", "/t", "REG_SZ", "/d", manifestPath, "/f").Run()
		}
		log += "✅ Registry Windows: OK\n"
		
		files, _ := ioutil.ReadDir("C:\\")
		for _, f := range files {
			if f.IsDir() {
				fullPath := filepath.Join("C:\\", f.Name())
				if injectWinFolder(fullPath, binPath) {
					log += fmt.Sprintf("✅ Auto-Inject Portable: %s\n", fullPath)
				}
			}
		}
	} else {
		if injectWinFolder(customPath, binPath) {
			log += fmt.Sprintf("✅ Custom Inject: %s\n", customPath)
		} else {
			log += "❌ Gagal: Folder tidak valid (Cari folder induk yang berisi Data/profile)\n"
		}
	}
	return log
}

func injectWinFolder(rootPath, binPath string) bool {
	if _, err := os.Stat(filepath.Join(rootPath, "Data", "profile")); os.IsNotExist(err) { return false }
	targetDir := filepath.Join(rootPath, "Data", "profile", "NativeMessagingHosts")
	os.MkdirAll(targetDir, 0755)
	manifestPath := filepath.Join(targetDir, HOST_NAME+".json")
	return createManifest(manifestPath, binPath) == nil
}
