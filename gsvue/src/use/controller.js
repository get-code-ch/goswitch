import {reactive, toRefs} from "vue";

export default function useController() {

    // const url = "ws://localhost:4444/ws";
    const url = "wss://get-code.ch:4444/ws";

    let connection;
    let id = makeId(5);
    let client;

    const controller = reactive({
        msg:"",
        status:"",
        deviceId:"",
        devices: null,
        switches: null,
    })

    function newConnection() {
        let obj;
        client = {"Type": "browser", "Id": "vue-" + id};
        connection = new WebSocket(url);

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
                    controller.status = "Server->" + obj.server.Id;
                    connection.send(JSON.stringify({
                        "client": client,
                        "action": "List",
                        "server": obj.server,
                        "data": ""
                    }))
                    break;
                case "acknowledge":
                    controller.status = "Acknowledgment";
                    controller.msg = obj.data;
                    break;
                case "gpiostate":
                    controller.status = "GPIO State received";
                    controller.msg = obj.data;
                    if (controller.switches != null) {
                        controller.switches.forEach(swc => {
                            if (swc.address == obj.data.address && swc.gpio == obj.data.gpio) {
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
                        "data": client
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
        console.log("Connection open -> ", event.data)
    }

    function connectionOnClose(event) {
        console.log("Socket is closed, Reconnect in 5 seconds -> ", event.data);
        connection = null;
        controller.devices = null;
        controller.switches = null;
        setTimeout(() =>{
            newConnection();
        }, 5000);
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
        let data = {"Action": "SetGPIO", "client": device, "data": i2c};

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
    return { ...toRefs(controller), newConnection, deviceInfo, toggleGpio };

}
