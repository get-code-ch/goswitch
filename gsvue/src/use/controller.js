import {reactive, toRefs} from "vue";

export default function useController() {

    let connection;
    let client;

    const controller = reactive({
        msg:"",
        status:"Hello",
        //deviceId:""
    })

    function newConnection(url) {
        let obj;
        connection = new WebSocket(url);
        client = {"Type": "browser", "Id": "vue-xxx"};
        connection.onmessage = function (event) {
            console.log(event.data);
            controller.msg = event.data;
            obj = JSON.parse(event.data);
            switch (obj.action.toLowerCase()) {
                case "register":
                    connection.send(JSON.stringify({
                        "client": client,
                        "action": "Register",
                        "server": obj.server,
                        "data": client
                    }));
                    break;
                case "accept":
                    controller.status = "Server->" + obj.server.Id;
                    connection.send(JSON.stringify({
                        "client": client,
                        "action": "List",
                        "server": obj.server,
                        "data": ""
                    }))
                    break;
                case "list":
                    controller.status = "Device list returned";
                    controller.msg = obj.data;
                    break;
                case "receiveinfo":
                    controller.status = "Device info received";
                    controller.msg = obj.data;
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

    return { ...toRefs(controller), newConnection, deviceInfo };

}
