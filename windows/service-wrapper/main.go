package main

// Windows Service Wrapper for ROI Agent
// This runs roi-agent.exe as a Windows Service

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const serviceName = "ROI Agent"

type roiAgentService struct {
	agentCmd *exec.Cmd
}

func (s *roiAgentService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}

	// Get executable directory
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		return true, 1
	}

	baseDir := filepath.Dir(exePath)
	agentPath := filepath.Join(baseDir, "roi-agent.exe")

	// Start roi-agent.exe
	s.agentCmd = exec.Command(agentPath)
	s.agentCmd.Dir = baseDir

	// Setup logging
	logDir := filepath.Join(baseDir, "logs")
	os.MkdirAll(logDir, 0755)
	logFile, err := os.OpenFile(filepath.Join(logDir, "service.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		s.agentCmd.Stdout = logFile
		s.agentCmd.Stderr = logFile
		defer logFile.Close()
	}

	if err := s.agentCmd.Start(); err != nil {
		log.Printf("Failed to start roi-agent: %v", err)
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Service loop
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			default:
				log.Printf("Unexpected service control request: %v", c)
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}

	// Stop roi-agent
	if s.agentCmd != nil && s.agentCmd.Process != nil {
		s.agentCmd.Process.Kill()
		s.agentCmd.Wait()
	}

	return false, 0
}

func runService() {
	elog, err := eventlog.Open(serviceName)
	if err != nil {
		return
	}
	defer elog.Close()

	elog.Info(1, "ROI Agent service starting")

	err = svc.Run(serviceName, &roiAgentService{})
	if err != nil {
		elog.Error(1, err.Error())
		return
	}

	elog.Info(1, "ROI Agent service stopped")
}

func main() {
	// Check if running as service
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatal(err)
	}

	if isService {
		runService()
		return
	}

	// Running as console app (for testing)
	log.Println("ROI Agent Service Wrapper")
	log.Println("This should be run as a Windows Service")
	log.Println("Use: sc create \"ROI Agent\" binPath= \"path\\to\\service-wrapper.exe\"")

	time.Sleep(5 * time.Second)
}
