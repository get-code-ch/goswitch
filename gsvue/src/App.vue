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
    <!--
    <input v-model="deviceId" hidden>
    <button @click="deviceInfo(deviceId)">Device Info</button>
    -->

    <div v-if="modules != null">
        Selected device: {{deviceId}}
        <ul>
            <li v-for="module in modules" :key="module.name">
                {{ module.name }} - {{ module.description }}
                <ul>
                    <li v-for="(gpio, index) in module.gpios" :key="index">
                        gpio {{ index }} - state {{ gpio}}
                        <button @click="toggleGpio(deviceId, module.name, index)">Toogle state</button>
                    </li>
                </ul>
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

            const {newConnection, deviceInfo, toggleGpio, msg, status, modules, devices, deviceId} = useController();
            return {msg, status, modules, devices, deviceId, deviceInfo, toggleGpio};
        }
    };
</script>