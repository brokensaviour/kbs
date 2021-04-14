package main

import (
  "log"
  "net"
  "os"
  "os/exec"
  "os/signal"
  "strconv"
  "strings"
  "syscall"
  "time"

  "github.com/MarinX/keylogger"
  "github.com/spf13/viper"
)

type config struct {
  BigArea    uint8
  MiddleArea uint8
  SmallArea  uint8
}
var cfg config

const (
  KFH = "/dev/console"
  LED_SCR = 0x1
  LED_NUM = 0x2
  LED_CAP = 0x4
  LED_OFFS = 0x0
  KDSETLED = 0x4b32
)
func locationAt() [3]int {
	return [3]int{LED_NUM, LED_CAP, LED_SCR}
}
var h int

func setLamps(lamps int) {
  _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(h), uintptr(KDSETLED), uintptr(lamps))
  if err != 0 {
    panic(err)
  }
}

func keyGrubberWithSender() {
  me, err := net.ListenPacket("udp4", ":9000")
  if err != nil {
    panic(err)
  }
  defer me.Close()
  go broadcastWatcher(me)
  broadcast, err := net.ResolveUDPAddr("udp4", "10.255.255.255:9000")
  if err != nil {
    panic(err)
  }
  keyboard := keylogger.FindKeyboardDevice()
  if len(keyboard) <= 0 {
    panic("No keyboard found..")
  }
  k, err := keylogger.New(keyboard)
  if err != nil {
    panic(err)
  }
  defer k.Close()
  events := k.Read()
  for e := range events {
    switch e.Type {
    case keylogger.EvKey:
      if e.KeyPress() {
        log.Println("[event] press key ", e.KeyString())
        _, err := me.WriteTo([]byte("saviour our souls"), broadcast)
        if err != nil {
          panic(err)
        }
      }
      break
    }
  }
}

func broadcastWatcher(pc net.PacketConn) {
  buf := make([]byte, 1024)
  for {
    _, addr, err := pc.ReadFrom(buf)
    if err != nil {
      panic(err)
    }
    b := addr.(*net.UDPAddr).IP.To4()
    inform(b)
  }
}

func inform(location net.IP) {
  for i := 1; i <= len(location)-1; i++ {
    for j := 0; j < int(location[i]); j++ {
      setLamps(locationAt()[i-1]) 
      time.Sleep(350 * time.Millisecond)
      setLamps(LED_OFFS) 
      time.Sleep(350 * time.Millisecond)
    }
    time.Sleep(2000 * time.Millisecond)
  }
}

func main() { 
  viper.SetConfigName("config")
  viper.AddConfigPath(".")
  err := viper.ReadInConfig()
  if err != nil {
    panic(err)
  }
  err = viper.Unmarshal(&cfg)
  if err != nil {
    panic(err)
  }
  if (cfg.BigArea > 9) || (cfg.MiddleArea > 9) || (cfg.SmallArea > 9) {
    panic("Numbers bigger 10 not permit for this version.")
  }
  cmd, err := exec.Command("/bin/bash", "batman_begin.sh").Output()
  if err != nil {
    panic(string(cmd))
  }
  ip := []string{"sudo ip addr add 10.",
                 strconv.Itoa(int(cfg.BigArea)), ".",
                 strconv.Itoa(int(cfg.MiddleArea)), ".",
                 strconv.Itoa(int(cfg.SmallArea)),
                 "/8 dev bat0"}
  res := strings.Join(ip, "")
  args := []string{"-c", res}
  _, err = exec.Command("/bin/bash", args...).Output()
  if err != nil {
    panic(err)
  }

  h, err = syscall.Open(KFH, os.O_RDONLY|syscall.O_CLOEXEC, 0666)
  if err != nil {
          panic(err)
  }
  defer syscall.Close(h)
 
  go keyGrubberWithSender()

  quit := make(chan os.Signal, 1)
  signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
  defer signal.Stop(quit)
  for {
    select {
    case <-quit:
      return
    }
  }
}
