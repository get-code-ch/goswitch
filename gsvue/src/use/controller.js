import {reactive, toRefs} from "vue";

export default function useController() {

    let connection;
    let client;
    let api_key;
    let reject = false;

    const controller = reactive({
        msg:"",
        status:"",
        deviceId:"",
        devices: null,
        switches: null,
        connected: false
    })

    function genApiKey() {
        let key = makeId(30);
        let clientId = "vue-" + makeId(5);
        localStorage.setItem("api_key", key);
        localStorage.setItem("client_id", clientId);
    }

    function newConnection() {
        let obj;
        api_key = localStorage.getItem("api_key");
        // client = {"Type": "browser", "Id": "vue-" + id};
        client = {"Type": "browser", "Id": localStorage.getItem("client_id")};
        connection = new WebSocket(genWsUrl());

        connection.onerror = function (event) {
            connectionOnError(event);
        }

        connection.onopen = function(event) {
            connectionOnOpen(event);
        }

        connection.onclose = function(event) {
            connectionOnClose(event);
        }

        connection.onmessage = function (event) {
            console.log(event.data);
            controller.msg = event.data;
            obj = JSON.parse(event.data);
            switch (obj.action.toLowerCase()) {
                case "accept":
                    reject = false;
                    controller.status = "Server->" + obj.server.Id;
                    browserNotify("Connection accepted");
                    connection.send(JSON.stringify({
                        "client": client,
                        "action": "List",
                        "server": obj.server,
                        "data": ""
                    }))
                    break;
                case "reject":
                    reject = true;
                    controller.status = "Rejected " + obj.data;
                    controller.msg = obj.data;
                    connection.close();
                    break;
                case "acknowledge":
                    controller.status = "Acknowledgment";
                    controller.msg = obj.data;
                    break;
                case "error":
                    controller.status = "Error:" + obj.data;
                    controller.msg = obj.data;
                    break;
                case "gpiostate":
                    controller.status = "GPIO State received";
                    controller.msg = obj.data;
                    if (controller.switches != null) {
                        controller.switches.forEach(swc => {
                            if (swc.mac_addr == obj.data.mac_addr && swc.address == obj.data.address && swc.gpio == obj.data.gpio) {
                                swc.state = obj.data.state;
                            }
                        });
                    }
                    break;
                case "list":
                    controller.status = "Device list returned";
                    controller.msg = obj.data;
                    controller.devices = obj.data;
                    if (controller.devices.findIndex((d) => d == controller.deviceId) == -1) {
                     controller.deviceId = "";
                     controller.switches = null;
                    }

                    break;
                case "receiveinfo":
                    controller.status = "Device info received";
                    controller.msg = obj.data;
                    controller.switches = obj.data.device.switches;
                    break;
                case "register":
                    connection.send(JSON.stringify({
                        "client": client,
                        "action": "Register",
                        "server": obj.server,
                        "data": {"client": client, "api_key": api_key}
                    }));
                    break;
                default:
                    console.log("Nothing to do for action : " + obj.action)
                    console.log("Received data : " + obj.data)
                    controller.status = "Nothing to do for action : " + obj.action
                    controller.msg = obj.data;
            }
        }

    }

    function connectionOnError(event) {
        console.log("Socket connection error -> ", event.data);
    }

    function connectionOnOpen(event) {
        controller.connected = 'connected';
        console.log("Connection open -> ", event.data)
    }

    function connectionOnClose(event) {
        controller.connected = 'disconnected';
        connection = null;
        controller.devices = null;
        controller.switches = null;
        if (!reject) {
            console.log("Socket is closed, Reconnect in 5 seconds -> ", event.data);
            setTimeout(() =>{
                newConnection();
            }, 5000);
        } else {
            console.log("Connection rejected by CommandCenter, do nothing");
        }
    }

    function deviceInfo(deviceId) {
        controller.deviceId = deviceId;
        let device = {"Type": "device", "Id": deviceId};
        let data = {"Action": "GetInfo",
            "data": client,
            "client": device
        };
        connection.send(JSON.stringify({
            "client": device,
            "action": "Relay",
            "data": data
        }));
        console.log(deviceId)
    }


    function toggleGpio(deviceId, address, gpio ) {
        let device = {"Type": "device", "Id": deviceId};

        let i2c = {"command":"reverse", "id": deviceId, "address": "" + address, "gpio": "" + gpio};
        let data = {"Action": "SetGPIO", "client": device, "data": {"i2c": i2c, "client": client}};

        connection.send(JSON.stringify({
            "action": "Relay",
            "device": device,
            "data": data
        }))
    }

    function makeId(length) {
        let result           = '';
        let characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let charactersLength = characters.length;
        for ( let i = 0; i < length; i++ ) {
            result += characters.charAt(Math.floor(Math.random() * charactersLength));
        }
        return result;
    }

    function genWsUrl() {
        let host = window.location.hostname;
        let protocol = location.protocol
        let port = location.port;

        if (port == 8081 || port == 8080) {
            port = "4433";
        }
        return protocol.toLowerCase() == "https:" ? "wss://" + host + ":" + port + "/ws" : "ws://" + host + ":" + port  + "/ws";

    }

    function browserNotify(msg) {
        // eslint-disable-next-line
        let notification = null;
        let isMobile = navigator.userAgent.match(/(iPhone|iPod|iPad|Android|webOS|BlackBerry|IEMobile|Opera Mini)/i)

         if (!("Notification" in window) || isMobile) {
            return ;
        }

        if (Notification.permission === "granted") {
            notification = new Notification(msg);
        } else {
            Notification.requestPermission().then(function(permission) {
                if (permission === "granted") {
                   notification = new Notification(msg);
                }
            })
        }
    }

    return { ...toRefs(controller), genApiKey, newConnection, deviceInfo, toggleGpio };

}
