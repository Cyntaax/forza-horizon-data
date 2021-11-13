package main

import (
	"encoding/binary"
	"fmt"
	"github.com/fatih/structtag"
	"log"
	"math"
	"net"
	"reflect"
	"strconv"
)

// FHDataPacket describes the data packet coming from Forza Horizon
//
// The tag "offset" describes where to start reading the buffer for the value
//
// The tag "length" describes how many bytes to get from the buffer for the value. defaults to `4`
type FHDataPacket struct {
	IsRaceOn                 bool    `offset:"0"`
	TimestampMs              uint    `offset:"4"`
	EngineMaxRpm             float32 `offset:"8"`
	EngineIdleRpm            float32 `offset:"12"`
	CurrentEngineRpm         float32 `offset:"16"`
	AccelerationX            float32 `offset:"20"`
	AccelerationY            float32 `offset:"24"`
	AccelerationZ            float32 `offset:"28"`
	VelocityX                float32 `offset:"32"`
	VelocityY                float32 `offset:"36"`
	VelocityZ                float32 `offset:"40"`
	AngularVelocityX         float32 `offset:"44"`
	AngularVelocityY         float32 `offset:"48"`
	AngularVelocityZ         float32 `offset:"52"`
	Yaw                      float32 `offset:"56"`
	Pitch                    float32 `offset:"60"`
	Roll                     float32 `offset:"64"`
	NormSuspensionTravelFl   float32 `offset:"68"`
	NormSuspensionTravelFr   float32 `offset:"72"`
	NormSuspensionTravelRl   float32 `offset:"76"`
	NormSuspensionTravelRr   float32 `offset:"80"`
	TireSlipRatioFl          float32 `offset:"84"`
	TireSlipRatioFr          float32 `offset:"88"`
	TireSlipRatioRl          float32 `offset:"92"`
	TireSlipRatioRr          float32 `offset:"96"`
	WheelRotationSpeedFl     float32 `offset:"100"`
	WheelRotationSpeedFr     float32 `offset:"104"`
	WheelRotationSpeedRl     float32 `offset:"108"`
	WheelRotationSpeedRr     float32 `offset:"112"`
	WheelOnRumbleStripFl     float32 `offset:"116"`
	WheelOnRumbleStripFr     float32 `offset:"120"`
	WheelOnRumbleStripRl     float32 `offset:"124"`
	WheelOnRumbleStripRr     float32 `offset:"128"`
	WheelInPuddleFl          float32 `offset:"132"`
	WheelInPuddleFr          float32 `offset:"136"`
	WheelInPuddleRl          float32 `offset:"140"`
	WheelInPuddleRr          float32 `offset:"144"`
	SurfaceRumbleFl          float32 `offset:"148"`
	SurfaceRumbleFr          float32 `offset:"152"`
	SurfaceRumbleRl          float32 `offset:"156"`
	SurfaceRumbleRr          float32 `offset:"160"`
	TireSlipAngleFl          float32 `offset:"164"`
	TireSlipAngleFr          float32 `offset:"168"`
	TireSlipAngleRl          float32 `offset:"172"`
	TireSlipAngleRr          float32 `offset:"176"`
	TireCombinedSlipFl       float32 `offset:"180"`
	TireCombinedSlipFr       float32 `offset:"184"`
	TireCombinedSlipRl       float32 `offset:"188"`
	TireCombinedSlipRr       float32 `offset:"192"`
	SuspensionTravelMetersFl float32 `offset:"196"`
	SuspensionTravelMetersFr float32 `offset:"200"`
	SuspensionTravelMetersRl float32 `offset:"204"`
	SuspensionTravelMetersRr float32 `offset:"208"`
	CarOrdinal               uint    `offset:"212" length:"1"`
	CarClass                 uint    `offset:"216" length:"1"`
	CarPerformanceIndex      uint    `offset:"220" length:"1"`
	DriveTrain               uint    `offset:"224" length:"1"`
	NumCylinders             uint    `offset:"228" length:"1"`
	PositionX                float32 `offset:"244"`
	PositionY                float32 `offset:"248"`
	PositionZ                float32 `offset:"252"`
	Speed                    float32 `offset:"256"`
	Power                    float32 `offset:"260"`
	Torque                   float32 `offset:"264"`
	TireTempFl               float32 `offset:"268"`
	TireTempFr               float32 `offset:"272"`
	TireTempRl               float32 `offset:"276"`
	TireTempRr               float32 `offset:"280"`
	Boost                    float32 `offset:"284"`
	Fuel                     float32 `offset:"288"`
	Distance                 float32 `offset:"292"`
	BestLapTime              float32 `offset:"296"`
	LastLapTime              float32 `offset:"300"`
	CurrentLapTime           float32 `offset:"304"`
	CurrentRaceTime          float32 `offset:"308"`
	Lap                      uint    `offset:"312" length:"2"`
	RacePosition             uint    `offset:"314" length:"1"`
	Accelerator              uint    `offset:"315" length:"1"`
	Brake                    uint    `offset:"316" length:"1"`
	Clutch                   uint    `offset:"317" length:"1"`
	Handbrake                uint    `offset:"318" length:"1"`
	Gear                     uint    `offset:"319" length:"1"`
	Steer                    int     `offset:"320" length:"1"`
	NormalDrivingLine        uint    `offset:"321" length:"1"`
	NormalAiBrakeDifference  uint    `offset:"322" length:"1"`
}

func main() {
	// infinite loop here of reading the data
	for {
		var data FHDataPacket
		bytes := readStream()
		unmarshallFHData(bytes, &data)
		fmt.Println("rpms", data.CurrentEngineRpm)
	}

}

// this function uses the struct tags to assign data from the buffer correctly to the corresponding fields
func unmarshallFHData(data []byte, value interface{}) {
	tmpValue := reflect.ValueOf(value)
	tmpValue = tmpValue.Elem()

	extractedType := tmpValue.Type()

	numFields := extractedType.NumField()

	for i := 0; i < numFields; i++ {
		tag := extractedType.Field(i).Tag
		tags, _ := structtag.Parse(string(tag))

		offset := 0
		length := 4

		offsetTag, err := tags.Get("offset")
		if err == nil {
			offset, _ = strconv.Atoi(offsetTag.Value())
		}

		lengthTag, err := tags.Get("length")
		if err == nil {
			length, _ = strconv.Atoi(lengthTag.Value())
		}

		outBytes := data[offset : offset+length]
		tp := extractedType.Field(i).Type.String()
		if length == 1 {
			if tp == "uint" {
				tmpValue.Field(i).Set(reflect.ValueOf(uint(outBytes[0])))
			} else if tp == "int" {
				tmpValue.Field(i).Set(reflect.ValueOf(int(outBytes[0])))
			}

		} else if length == 2 {
			bits := binary.LittleEndian.Uint16(outBytes)
			tmpValue.Field(i).Set(reflect.ValueOf(uint(bits)))
		} else {

			bits := binary.LittleEndian.Uint32(outBytes)
			float := math.Float32frombits(bits)

			switch tp {
			case "uint":
				tmpValue.Field(i).Set(reflect.ValueOf(uint(float)))
			case "float32":
				tmpValue.Field(i).Set(reflect.ValueOf(float))
			case "bool":
				intf := int(float)
				if intf == 0 {
					tmpValue.Field(i).Set(reflect.ValueOf(false))
				} else {
					tmpValue.Field(i).Set(reflect.ValueOf(true))
				}
			case "int":
				tmpValue.Field(i).Set(reflect.ValueOf(int(float)))
			}
		}
	}
}

// reads the udp stream and returns the buffer as []byte
func readStream() []byte {

	CONNECT := "0.0.0.0:9999"

	s, err := net.ResolveUDPAddr("udp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	c, err := net.ListenUDP("udp", s)
	if err != nil {
		fmt.Println(err)
		return []byte{}
	}
	defer c.Close()

	buffer := make([]byte, 1500)

	_, addr, err := c.ReadFromUDP(buffer)
	if err != nil {
		log.Fatal("Error reading UDP data:", err, addr)
	}

	return buffer
}
