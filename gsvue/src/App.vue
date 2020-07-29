<template>
    <div v-if="devices != null">
        <div>
            <p v-for="device in devices" :key="device.Id">
                <button @click="deviceInfo(device)">{{ device }}</button>
            </p>
        </div>
    </div>

    <div v-if="switches != null">
        Selected device: {{deviceId}}
        <div>
            <p v-for="swc in switches" :key="swc.name">
                <button class="btn" v-bind:class="[swc.state == 0 ? 'off' : 'on']" @click="toggleGpio(deviceId, swc.address, swc.gpio)">{{swc.name}}</button>
            </p>
        </div>
    </div>
    <div>
        <p>status: {{status}}</p>
<!--        <p>msg: {{msg}}</p>-->
    </div>

</template>
<script>
    import useController from "./use/controller";
    import {onMounted} from "vue"

    export default {
        setup() {
            onMounted(() => {
                newConnection();
                console.log("Mounted");
            })

            const {newConnection, deviceInfo, toggleGpio, msg, status, deviceId, devices, switches } = useController();
            return {msg, status, deviceId, devices, switches, deviceInfo, toggleGpio};
        }
    };
</script>