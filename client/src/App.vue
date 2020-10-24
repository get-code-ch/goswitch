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

  <!-- Presents endpoints - action or values -->
  <div v-if="ICs != null">
    <div class="state-btn-list">
      <ul>
        <div v-for="(ic, key) in ICs" :key="ic.address">
          <li v-for="ep in ic.endPoints" :key="ep.id">
            <button class="state-btn" v-bind:class="[ep.attributes.state == 0 ? 'off' : 'on']"
                    @click="btnClick(ic.type, key, ep.id, ep.attributes)">
              <!-- {{ key }}|{{ ep.id }} - -->{{ ep.name }}
            </button>
          </li>
        </div>
      </ul>
    </div>
  </div>

  <!--
  <div>
    <graph v-model:g-props="graphProperties"></graph>
  </div>
  -->

</template>
<script>
import useController from "./use/controller";
//import graph from "./components/graph"
import {onMounted} from "vue"

export default {
  setup() {
    onMounted(() => {
      //navigator.serviceWorker.register()
      if (localStorage.getItem("api_key") == null) {
        genApiKey();
      }
      newConnection();

    })

    const {genApiKey, newConnection, deviceInfo, toggleGpio, btnClick, msg, status, deviceId, devices, ICs, connected, graphProperties} = useController();
    return {
      msg,
      status,
      deviceId,
      devices,
      ICs,
      deviceInfo,
      toggleGpio,
      btnClick,
      connected,
      graphProperties
    };
  },
  components: {
    //  graph
  }
};
</script>