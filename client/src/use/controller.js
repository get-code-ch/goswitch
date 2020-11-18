import {reactive, toRefs} from "vue";

export default function useController() {

    let connection;
    let client;
    let api_key;
    let reject = false;

    const controller = reactive({
        msg: "",
        status: "",
        deviceId: "",
        devices: null,
        ICs: null,
        connected: false,
        graphProperties: {
            width: 1000,
            height: 400,
            title: "My Graph",
            data: drawLines(100)
        }

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

        connection.onopen = function (event) {
            connectionOnOpen(event);
        }

        connection.onclose = function (event) {
            connectionOnClose(event);
        }

        connection.onmessage = function (event) {
            let device = null;
            let data = null;
            let hl = 0;

            //console.log(event);
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
                    // Loop all gpio state and update status on
                    obj.data.forEach(ep => {
                        let icepIdx = controller.ICs[ep.address].endPoints.findIndex((icep) => icep.id == ep.id)
                        if (icepIdx >= 0) {
                            controller.ICs[ep.address].endPoints[icepIdx].attributes.state = ep.state
                        }
                    })
                    break;
                case "digitalvalue":

                    controller.status = "New digital value received";

                    obj.data.forEach(ep => {
                        let icepIdx = controller.ICs[ep.address].endPoints.findIndex((icep) => icep.id == ep.id)
                        if (icepIdx >= 0) {
                            controller.ICs[ep.address].endPoints[icepIdx].attributes.value = ep.value

                            // Add value to history
                            if (!Object.prototype.hasOwnProperty.call(controller.ICs[ep.address].endPoints[icepIdx], "history")) {
                                controller.ICs[ep.address].endPoints[icepIdx].history = new Array(1);
                                controller.ICs[ep.address].endPoints[icepIdx].history[0] = controller.ICs[ep.address].endPoints[icepIdx].attributes.value
                            } else {
                                hl = controller.ICs[ep.address].endPoints[icepIdx].history.length;
                                if (hl <= 10) {
                                    controller.ICs[ep.address].endPoints[icepIdx].history.push(controller.ICs[ep.address].endPoints[icepIdx].attributes.value);
                                } else {
                                    controller.ICs[ep.address].endPoints[icepIdx].history.push(controller.ICs[ep.address].endPoints[icepIdx].attributes.value);
                                    controller.ICs[ep.address].endPoints[icepIdx].history.shift();
                                }
                            }
                        }
                    })


                    break;
                case "list":
                    controller.status = "Device list returned";
                    controller.msg = obj.data;
                    controller.devices = obj.data;
                    if (controller.devices.findIndex((d) => d == controller.deviceId) == -1) {
                        controller.deviceId = "";
                    }

                    break;
                case "receiveinfo":
                    controller.status = "Device info received";
                    controller.msg = obj.data;

                    if ('endPoints' in obj.data) {
                        if (controller.ICs == null) {
                            controller.ICs = {}
                        }
                        controller.ICs[obj.data.address] = {}
                        controller.ICs[obj.data.address].endPoints = obj.data.endPoints
                        controller.ICs[obj.data.address].type = obj.data.type
                        //controller.ICs[obj.data.address].history = [];
                    }

                    if (controller.msg.type == "mcp23008") {
                        device = {"Type": "device", "Id": controller.deviceId};
                        data = {
                            "Action": "GetAllGPIOState",
                            "data": client,
                            "client": device
                        };
                        connection.send(JSON.stringify({
                            "client": device,
                            "action": "Relay",
                            "data": data
                        }));
                    }
                    if (controller.msg.type == "ads1115") {
                        device = {"Type": "device", "Id": controller.deviceId};
                        data = {
                            "Action": "GetAllValues",
                            "data": client,
                            "client": device
                        };
                        connection.send(JSON.stringify({
                            "client": device,
                            "action": "Relay",
                            "data": data
                        }));
                    }
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
        controller.ICs = null;
        if (!reject) {
            console.log("Socket is closed, Reconnect in 5 seconds -> ", event.data);
            setTimeout(() => {
                newConnection();
            }, 5000);
        } else {
            console.log("Connection rejected by CommandCenter, do nothing");
        }
    }

    function deviceInfo(deviceId) {
        controller.deviceId = deviceId;
        let device = {"Type": "device", "Id": deviceId};
        let data = {
            "Action": "GetInfo",
            "data": client,
            "client": device
        };
        connection.send(JSON.stringify({
            "client": device,
            "action": "Relay",
            "data": data
        }));
    }


    function toggleGpio(deviceId, address, gpio) {
        let device = {"Type": "device", "Id": deviceId};

        let i2c = {"command": "reverse", "id": deviceId, "address": "" + address, "gpio": "" + gpio};
        let data = {"Action": "SetGPIO", "client": device, "data": {"i2c": i2c, "client": client}};

        connection.send(JSON.stringify({
            "action": "Relay",
            "device": device,
            "data": data
        }))
    }

    function btnClickMCP23008(deviceId, type, key, id, attributes) {
        let device = null;
        let data = null;
        if (attributes.mode.toLowerCase() === "output" || attributes.mode.toLowerCase() == "push") {
            attributes.state = Math.abs(attributes.state - 1)
            device = {"Type": "device", "Id": deviceId};
            data = {
                "Action": "SetGPIO",
                "client": device,
                "data": {"client": client, "address": key, "id": id, "attributes": attributes}
            };

            controller.status = "Btn " + key + " " + id + " clicked!"
            connection.send(JSON.stringify({
                "action": "Relay",
                "device": device,
                "data": data
            }));
        } else {
            controller.status = "Input mode for mcp23008, No allowed action";
        }
    }

    function btnClickADS1115(deviceId, type, key, id, attributes) {
        let device = null;
        let data = null;
        device = {"Type": "device", "Id": controller.deviceId};
        data = {
            "Action": "GetValue",
            "client": device,
            "data": {"client": client, "address": key, "id": id, "attributes": attributes}
        };
        connection.send(JSON.stringify({
            "action": "Relay",
            "device": device,
            "data": data
        }));
    }

    function makeId(length) {
        let result = '';
        let characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
        let charactersLength = characters.length;
        for (let i = 0; i < length; i++) {
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
            host = "precision";
        }
        return protocol.toLowerCase() == "https:" ? "wss://" + host + ":" + port + "/ws" : "ws://" + host + ":" + port + "/ws";

    }

    function browserNotify(msg) {
        // eslint-disable-next-line
        let notification = null;
        let isMobile = navigator.userAgent.match(/(iPhone|iPod|iPad|Android|webOS|BlackBerry|IEMobile|Opera Mini)/i)

        if (!("Notification" in window) || isMobile) {
            return;
        }

        if (Notification.permission === "granted") {
            notification = new Notification(msg);
        } else {
            Notification.requestPermission().then(function (permission) {
                if (permission === "granted") {
                    notification = new Notification(msg);
                }
            })
        }
    }

    function drawLines(count) {
        let array = [];

        for (let i = 1; i < count; i++) {
            array.push({"x": i, "y": i * i});
        }
        return array
    }

    return {
        ...toRefs(controller),
        genApiKey,
        newConnection,
        deviceInfo,
        toggleGpio,
        btnClickADS1115,
        btnClickMCP23008,
        drawLines
    };

}
