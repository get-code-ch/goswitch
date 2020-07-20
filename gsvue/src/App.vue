<template>
    <p>msg: {{msg}}</p>
    <p>status: {{status}}</p>
    Device: <input v-model="deviceId">
    <button @click="deviceInfo(deviceId)">Device Info</button>


    <div v-if="modules != null">
        Selectect device: {{deviceId}}
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
    import {onMounted, ref} from "vue"

    export default {
        setup() {
            const deviceId = ref("")
            onMounted(() => {
                newConnection("ws://localhost:4444/ws");
                console.log("Mounted");
            })

            const {newConnection, deviceInfo, toggleGpio, msg, status, modules} = useController();
            return {msg, status, modules, deviceId, deviceInfo, toggleGpio};
        }
    };
</script>