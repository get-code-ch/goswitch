import {reactive, toRefs} from "vue";

export default function useController() {

    let connection;
    let client;

    const controller = reactive({
        msg:"",
        status:"",
        deviceId:"",
        devices: null,
        modules: null,
    })

    function newConnection(url) {
        let obj;
        connection = new WebSocket(url);
        client = {"Type": "browser", "Id": "vue-" + makeid(5)};
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
                case "list":
                    controller.status = "Device list returned";
                    controller.msg = obj.data;
                    controller.devices = obj.data;
                    break;
                case "receiveinfo":
                    controller.status = "Device info received";
                    controller.msg = obj.data;
                    controller.modules = obj.data.device.Modules;
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


    function toggleGpio(deviceId, name, gpio ) {
        let device = {"Type": "device", "Id": deviceId};

        let i2c = {"command":"reverse", "id": deviceId, "module": name, "sw": "" + gpio};
        let data = {"Action": "SetGPIO", "client": device, "data": i2c};

        connection.send(JSON.stringify({
            "action": "Relay",
            "device": device,
            "data": data
        }))
    }


    function makeid(length) {
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
