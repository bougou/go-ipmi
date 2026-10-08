// The installed target Linux headers are the ABI oracle. No IPMI device is used.
#include <sys/ioctl.h>
#include <linux/ipmi.h>
#include <stddef.h>
#include <stdio.h>
#define VALUE(name, value) printf(name " %llu\n", (unsigned long long)(value))
#define LAYOUT(t) VALUE(#t ".size", sizeof(struct t)); VALUE(#t ".align", _Alignof(struct t))
#define FIELD(t, f) \
 VALUE(#t "." #f ".offset", offsetof(struct t, f)); \
 VALUE(#t "." #f ".size", sizeof(((struct t *)0)->f)); \
 VALUE(#t "." #f ".signed", _Generic(((struct t *)0)->f, signed char: 1, short: 1, int: 1, long: 1, long long: 1, default: 0))
int main(void) {
    LAYOUT(ipmi_addr);
    FIELD(ipmi_addr, addr_type);
    FIELD(ipmi_addr, channel);
    FIELD(ipmi_addr, data);
    LAYOUT(ipmi_system_interface_addr);
    FIELD(ipmi_system_interface_addr, addr_type);
    FIELD(ipmi_system_interface_addr, channel);
    FIELD(ipmi_system_interface_addr, lun);
    LAYOUT(ipmi_ipmb_addr);
    FIELD(ipmi_ipmb_addr, addr_type);
    FIELD(ipmi_ipmb_addr, channel);
    FIELD(ipmi_ipmb_addr, slave_addr);
    FIELD(ipmi_ipmb_addr, lun);
    LAYOUT(ipmi_ipmb_direct_addr);
    FIELD(ipmi_ipmb_direct_addr, addr_type);
    FIELD(ipmi_ipmb_direct_addr, channel);
    FIELD(ipmi_ipmb_direct_addr, slave_addr);
    FIELD(ipmi_ipmb_direct_addr, rs_lun);
    FIELD(ipmi_ipmb_direct_addr, rq_lun);
    LAYOUT(ipmi_lan_addr);
    FIELD(ipmi_lan_addr, addr_type);
    FIELD(ipmi_lan_addr, channel);
    FIELD(ipmi_lan_addr, privilege);
    FIELD(ipmi_lan_addr, session_handle);
    FIELD(ipmi_lan_addr, remote_SWID);
    FIELD(ipmi_lan_addr, local_SWID);
    FIELD(ipmi_lan_addr, lun);
    LAYOUT(ipmi_msg);
    FIELD(ipmi_msg, netfn);
    FIELD(ipmi_msg, cmd);
    FIELD(ipmi_msg, data_len);
    FIELD(ipmi_msg, data);
    LAYOUT(ipmi_req);
    FIELD(ipmi_req, addr);
    FIELD(ipmi_req, addr_len);
    FIELD(ipmi_req, msgid);
    FIELD(ipmi_req, msg);
    LAYOUT(ipmi_recv);
    FIELD(ipmi_recv, recv_type);
    FIELD(ipmi_recv, addr);
    FIELD(ipmi_recv, addr_len);
    FIELD(ipmi_recv, msgid);
    FIELD(ipmi_recv, msg);
    LAYOUT(ipmi_req_settime);
    FIELD(ipmi_req_settime, req);
    FIELD(ipmi_req_settime, retries);
    FIELD(ipmi_req_settime, retry_time_ms);
    LAYOUT(ipmi_cmdspec);
    FIELD(ipmi_cmdspec, netfn);
    FIELD(ipmi_cmdspec, cmd);
    LAYOUT(ipmi_cmdspec_chans);
    FIELD(ipmi_cmdspec_chans, netfn);
    FIELD(ipmi_cmdspec_chans, cmd);
    FIELD(ipmi_cmdspec_chans, chans);
    LAYOUT(ipmi_channel_lun_address_set);
    FIELD(ipmi_channel_lun_address_set, channel);
    FIELD(ipmi_channel_lun_address_set, value);
    LAYOUT(ipmi_timing_parms);
    FIELD(ipmi_timing_parms, retries);
    FIELD(ipmi_timing_parms, retry_time_ms);
    VALUE("IPMICTL_SEND_COMMAND", IPMICTL_SEND_COMMAND);
    VALUE("IPMICTL_SEND_COMMAND_SETTIME", IPMICTL_SEND_COMMAND_SETTIME);
    VALUE("IPMICTL_RECEIVE_MSG", IPMICTL_RECEIVE_MSG);
    VALUE("IPMICTL_RECEIVE_MSG_TRUNC", IPMICTL_RECEIVE_MSG_TRUNC);
    VALUE("IPMICTL_REGISTER_FOR_CMD", IPMICTL_REGISTER_FOR_CMD);
    VALUE("IPMICTL_UNREGISTER_FOR_CMD", IPMICTL_UNREGISTER_FOR_CMD);
    VALUE("IPMICTL_REGISTER_FOR_CMD_CHANS", IPMICTL_REGISTER_FOR_CMD_CHANS);
    VALUE("IPMICTL_UNREGISTER_FOR_CMD_CHANS", IPMICTL_UNREGISTER_FOR_CMD_CHANS);
    VALUE("IPMICTL_SET_GETS_EVENTS_CMD", IPMICTL_SET_GETS_EVENTS_CMD);
    VALUE("IPMICTL_SET_MY_CHANNEL_ADDRESS_CMD", IPMICTL_SET_MY_CHANNEL_ADDRESS_CMD);
    VALUE("IPMICTL_GET_MY_CHANNEL_ADDRESS_CMD", IPMICTL_GET_MY_CHANNEL_ADDRESS_CMD);
    VALUE("IPMICTL_SET_MY_CHANNEL_LUN_CMD", IPMICTL_SET_MY_CHANNEL_LUN_CMD);
    VALUE("IPMICTL_GET_MY_CHANNEL_LUN_CMD", IPMICTL_GET_MY_CHANNEL_LUN_CMD);
    VALUE("IPMICTL_SET_MY_ADDRESS_CMD", IPMICTL_SET_MY_ADDRESS_CMD);
    VALUE("IPMICTL_GET_MY_ADDRESS_CMD", IPMICTL_GET_MY_ADDRESS_CMD);
    VALUE("IPMICTL_SET_MY_LUN_CMD", IPMICTL_SET_MY_LUN_CMD);
    VALUE("IPMICTL_GET_MY_LUN_CMD", IPMICTL_GET_MY_LUN_CMD);
    VALUE("IPMICTL_SET_TIMING_PARMS_CMD", IPMICTL_SET_TIMING_PARMS_CMD);
    VALUE("IPMICTL_GET_TIMING_PARMS_CMD", IPMICTL_GET_TIMING_PARMS_CMD);
    VALUE("IPMICTL_GET_MAINTENANCE_MODE_CMD", IPMICTL_GET_MAINTENANCE_MODE_CMD);
    VALUE("IPMICTL_SET_MAINTENANCE_MODE_CMD", IPMICTL_SET_MAINTENANCE_MODE_CMD);
    VALUE("_IOC_NRBITS", _IOC_NRBITS);
    VALUE("_IOC_TYPEBITS", _IOC_TYPEBITS);
    VALUE("_IOC_SIZEBITS", _IOC_SIZEBITS);
    VALUE("_IOC_DIRBITS", _IOC_DIRBITS);
    VALUE("_IOC_NONE", _IOC_NONE);
    VALUE("_IOC_READ", _IOC_READ);
    VALUE("_IOC_WRITE", _IOC_WRITE);
    VALUE("_IOC_NRMASK", _IOC_NRMASK);
    VALUE("_IOC_TYPEMASK", _IOC_TYPEMASK);
    VALUE("_IOC_SIZEMASK", _IOC_SIZEMASK);
    VALUE("_IOC_DIRMASK", _IOC_DIRMASK);
    VALUE("_IOC_NRSHIFT", _IOC_NRSHIFT);
    VALUE("_IOC_TYPESHIFT", _IOC_TYPESHIFT);
    VALUE("_IOC_SIZESHIFT", _IOC_SIZESHIFT);
    VALUE("_IOC_DIRSHIFT", _IOC_DIRSHIFT);
    VALUE("IO", _IO(0x69, 0x55));
    VALUE("IOR", _IOR(0x69, 0x55, unsigned int));
    VALUE("IOW", _IOW(0x69, 0x55, unsigned int));
    VALUE("IOWR", _IOWR(0x69, 0x55, unsigned int));
    return 0;
}
