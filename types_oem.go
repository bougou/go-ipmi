package ipmi

// OEM represents Manufacturer ID, that is IANA Private Enterprise Number
type OEM uint32

// cSpell: disable
const (
	OEM_UNKNOWN          = 0
	OEM_DEBUG            = 0xFFFFFE /* Hoping IANA won't hit this soon */
	OEM_RESERVED         = 0x0FFFFF /* As per IPMI 2.0 specification */
	OEM_IBM_2            = 2        /* 2 for [IBM] */
	OEM_HP               = 11
	OEM_SUN              = 42
	OEM_NOKIA            = 94
	OEM_BULL             = 107
	OEM_HITACHI_116      = 116
	OEM_NEC              = 119
	OEM_TOSHIBA          = 186
	OEM_ERICSSON         = 193
	OEM_INTEL            = 343
	OEM_TATUNG           = 373
	OEM_HITACHI_399      = 399
	OEM_DELL             = 674
	OEM_HUAWEI           = 2011
	OEM_LMC              = 2168
	OEM_ASUSTEK          = 2623
	OEM_RADISYS          = 4337
	OEM_BROADCOM         = 4413
	OEM_IBM_4769         = 4769 /* 4769 for [IBM Corporation] */
	OEM_MAGNUM           = 5593
	OEM_TYAN             = 6653
	OEM_QUANTA           = 7244
	OEM_VIKING           = 9237
	OEM_ADVANTECH        = 10297
	OEM_FUJITSU_SIEMENS  = 10368
	OEM_AVOCENT          = 10418
	OEM_PEPPERCON        = 10437
	OEM_SUPERMICRO       = 10876
	OEM_OSA              = 11102
	OEM_GOOGLE           = 11129
	OEM_PICMG            = 12634
	OEM_RARITAN          = 13742
	OEM_KONTRON          = 15000
	OEM_PPS              = 16394
	OEM_IBM_20301        = 20301 /* 20301 for [IBM eServer X] */
	OEM_AMI              = 20974
	OEM_FOXCONN          = 22238
	OEM_ADLINK_24339     = 24339 /* 24339 for [ADLINK TECHNOLOGY INC.] */
	OEM_H3C              = 25506
	OEM_NOKIA_28458      = 28458
	OEM_VITA             = 33196
	OEM_INSPUR           = 37945
	OEM_TENCENT          = 41475
	OEM_BYTEDANCE        = 46045
	OEM_SUPERMICRO_47488 = 47488
	OEM_YADRO            = 49769
)

func (oem OEM) String() string {
	m := map[OEM]string{
		OEM_IBM_2:            "IBM",
		OEM_HP:               "HP",
		OEM_SUN:              "Sun",
		OEM_NOKIA:            "Nokia",
		OEM_BULL:             "Bull",
		OEM_HITACHI_116:      "Hitachi",
		OEM_NEC:              "NEC",
		OEM_TOSHIBA:          "Toshiba",
		OEM_ERICSSON:         "Ericsson",
		OEM_INTEL:            "Intel",
		OEM_TATUNG:           "Tatung", // 大同
		OEM_HITACHI_399:      "Hitachi",
		OEM_DELL:             "Dell",
		OEM_HUAWEI:           "Huawei",
		OEM_LMC:              "LMC",
		OEM_ASUSTEK:          "Asustek",
		OEM_RADISYS:          "Radisys",
		OEM_BROADCOM:         "Broadcom",
		OEM_IBM_4769:         "IBM",
		OEM_MAGNUM:           "Magnum", // 迈格纳技术集成公司
		OEM_TYAN:             "Tyan",   // 泰安
		OEM_QUANTA:           "Quanta",
		OEM_VIKING:           "Viking",
		OEM_ADVANTECH:        "Advantech", // 研华科技
		OEM_FUJITSU_SIEMENS:  "Fujitsu",
		OEM_AVOCENT:          "Avocent",
		OEM_PEPPERCON:        "Peppercon",
		OEM_SUPERMICRO:       "Supermicro",
		OEM_OSA:              "OSA",
		OEM_GOOGLE:           "Google",
		OEM_PICMG:            "PICMG",
		OEM_RARITAN:          "Raritan", // 力登
		OEM_KONTRON:          "Kontron", // 控创
		OEM_PPS:              "PPS",
		OEM_IBM_20301:        "IBM",
		OEM_AMI:              "AMI",
		OEM_FOXCONN:          "Foxconn",
		OEM_ADLINK_24339:     "ADLINK", // 凌华
		OEM_H3C:              "H3C",
		OEM_NOKIA_28458:      "Nokia",
		OEM_VITA:             "Vita",      // 维塔
		OEM_INSPUR:           "Inspur",    // 浪潮
		OEM_TENCENT:          "Tencent",   // 腾讯
		OEM_BYTEDANCE:        "ByteDance", // 字节跳动
		OEM_SUPERMICRO_47488: "Supermicro",
		OEM_YADRO:            "Yadro",
	}
	if s, ok := m[oem]; ok {
		return s
	}
	return "Unknown"
}
