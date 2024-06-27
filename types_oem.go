package ipmi

// OEM represents Manufacturer ID, that is IANA Private Enterprise Number
type OEM uint32

// cSpell: disable
const (
	OEM_UNKNOWN                      = 0
	OEM_DEBUG                        = 0xFFFFFE /* Hoping IANA won't hit this soon */
	OEM_RESERVED                     = 0x0FFFFF /* As per IPMI 2.0 specification */
	OEM_IBM_2                        = 2        /* 2 for [IBM] */
	OEM_HP                           = 11
	OEM_SUN                          = 42
	OEM_NOKIA                        = 94
	OEM_BULL                         = 107
	OEM_HITACHI_116                  = 116
	OEM_NEC                          = 119
	OEM_TOSHIBA                      = 186
	OEM_ERICSSON                     = 193
	OEM_INTEL                        = 343
	OEM_TATUNG                       = 373
	OEM_HITACHI_399                  = 399
	OEM_DELL                         = 674
	OEM_HUAWEI                       = 2011
	OEM_LMC                          = 2168
	OEM_RADISYS                      = 4337
	OEM_BROADCOM                     = 4413
	OEM_IBM_4769                     = 4769 /* 4769 for [IBM Corporation] */
	OEM_MAGNUM                       = 5593
	OEM_TYAN                         = 6653
	OEM_QUANTA                       = 7244
	OEM_VIKING                       = 9237
	OEM_ADVANTECH                    = 10297
	OEM_FUJITSU_SIEMENS              = 10368
	OEM_AVOCENT                      = 10418
	OEM_PEPPERCON                    = 10437
	OEM_SUPERMICRO                   = 10876
	OEM_OSA                          = 11102
	OEM_GOOGLE                       = 11129
	OEM_PICMG                        = 12634
	OEM_RARITAN                      = 13742
	OEM_KONTRON                      = 15000
	OEM_PPS                          = 16394
	OEM_IBM_20301                    = 20301 /* 20301 for [IBM eServer X] */
	OEM_AMI                          = 20974
	OEM_FOXCONN                      = 22238
	OEM_ADLINK_24339                 = 24339 /* 24339 for [ADLINK TECHNOLOGY INC.] */
	OEM_H3C                          = 25506
	OEM_NOKIA_SOLUTIONS_AND_NETWORKS = 28458
	OEM_VITA                         = 33196
	OEM_INSPUR                       = 37945
	OEM_TENCENT                      = 41475
	OEM_BYTEDANCE                    = 46045
	OEM_SUPERMICRO_47488             = 47488
	OEM_YADRO                        = 49769
)

func (oem OEM) String() string {
	m := map[OEM]string{
		2:     "IBM",
		11:    "HP",
		42:    "Sun",
		94:    "Nokia",
		107:   "Bull",
		116:   "Hitachi",
		119:   "NEC",
		186:   "Toshiba",
		193:   "Ericsson",
		343:   "Intel",
		373:   "Tatung", // 大同
		399:   "Hitachi",
		674:   "Dell",
		2011:  "Huawei",
		2168:  "LMC",
		4337:  "Radisys",
		4413:  "Broadcom",
		4769:  "IBM",
		5593:  "Magnum", // 迈格纳技术集成公司
		6653:  "Tyan",   // 泰安
		7244:  "Quanta",
		9237:  "Viking",
		10297: "Advantech", // 研华科技
		10368: "Fujitsu",
		10418: "Avocent",
		10437: "Peppercon",
		10876: "Supermicro",
		11102: "OSA",
		11129: "Google",
		12634: "PICMG",
		13742: "Raritan", // 力登
		15000: "Kontron", // 控创
		16394: "PPS",
		20301: "IBM",
		20974: "AMI",
		22238: "Foxconn",
		24339: "ADLINK", // 凌华
		25506: "H3C",
		28458: "Nokia",
		33196: "Vita",      // 维塔
		37945: "Inspur",    // 浪潮
		41475: "Tencent",   // 腾讯
		46045: "ByteDance", // 字节跳动
		47488: "Supermicro",
		49769: "Yadro",
	}
	if s, ok := m[oem]; ok {
		return s
	}
	return "Unknown"
}
