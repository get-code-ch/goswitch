<template>
    <p>msg: {{msg}}</p>
    <p>status: {{status}}</p>
    Device: <input v-model="deviceId">
    <button @click="deviceInfo(deviceId)">Device Info</button>
</template>
<script>
    import useController from "./use/controller";
    import {onMounted, ref} from "vue"

    export default {
        setup() {
            const deviceId = ref("")
            onMounted(() => {
                newConnection("ws://localhost:4444/ws");
                console.log("Mounted");
            })

            const {newConnection, deviceInfo, msg, status} = useController();
            return {msg, status, deviceId, deviceInfo};
        }
    };
</script>