import {reactive, toRefs} from "vue";

export default function useController() {
    const controller = reactive({
        msg:"",
        status:"Hello"
    })

    function newConnection(url) {
        let obj;
        let client;
        let connection = new WebSocket(url);
        connection.onmessage = function (event) {
            console.log(event.data);
            controller.msg = event.data;
            obj = JSON.parse(event.data);
            client = {"Type": "browser", "Id": "vue-xxx"};
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
                default:
                    console.log("Nothing to do for action : " + obj.action)
                    console.log("Received data : " + obj.data)
                    controller.status = "Nothing to do for action : " + obj.action
                    controller.msg = obj.data;
            }
        }

    }

    return { ...toRefs(controller), newConnection };

}
