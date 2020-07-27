<template>
    <p>msg: {{msg}}</p>
    <p>status: {{status}}</p>

    <div v-if="devices != null">
        <ul>
            <li v-for="device in devices" :key="device.Id">
                <button @click="deviceInfo(device)">{{ device }}</button>
            </li>
        </ul>
    </div>

    <div v-if="switches != null">
        Selected device: {{deviceId}}
        <ul>
            <li v-for="swc in switches" :key="swc.name">
                {{ swc.name }} - {{ swc.state }}
                <button @click="toggleGpio(deviceId, swc.address, swc.gpio)">{{swc.name}}</button>
            </li>
        </ul>
    </div>

</template>
<script>
    import useController from "./use/controller";
    import {onMounted} from "vue"

    export default {
        setup() {
            onMounted(() => {
                newConnection("ws://localhost:4444/ws");
                console.log("Mounted");
            })

            const {newConnection, deviceInfo, toggleGpio, msg, status, deviceId, devices, modules, switches } = useController();
            return {msg, status, deviceId, devices, modules, switches, deviceInfo, toggleGpio};
        }
    };
</script>