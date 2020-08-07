<template>
  <div>
    <p class="state" v-bind:class="connected">{{ connected }} {{ status }}</p>
  </div>

  <div v-if="devices != null">
    <div class="device-list">
      <ul>
        <li v-for="device in devices" :key="device.mac_addr">
          <button :disabled="!device.is_online" class="device-btn"
                  v-bind:class="[device.mac_addr == deviceId ? 'selected' : '']" @click="deviceInfo(device.mac_addr)">
            {{ device.name }}
          </button>
        </li>
      </ul>
    </div>
  </div>

  <div v-if="switches != null">
    <div class="state-btn-list">
      <ul>
        <li v-for="swc in switches" :key="swc.name">
          <button class="state-btn" v-bind:class="[swc.state == 0 ? 'off' : 'on']"
                  @click="toggleGpio(deviceId, swc.address, swc.gpio)">
            {{ swc.name }}
          </button>
        </li>
      </ul>
    </div>
  </div>

</template>
<script>
import useController from "./use/controller";
import {onMounted} from "vue"

export default {
  setup() {
    onMounted(() => {
      navigator.serviceWorker.register()
      if (localStorage.getItem("api_key") == null) {
        genApiKey();
      }
      newConnection();
    })

    const {genApiKey, newConnection, deviceInfo, toggleGpio, msg, status, deviceId, devices, switches, connected} = useController();
    return {msg, status, deviceId, devices, switches, deviceInfo, toggleGpio, connected};
  }
};
</script>