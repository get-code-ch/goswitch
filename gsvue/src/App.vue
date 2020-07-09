<template>
    <p>msg: {{msg}}</p>
    <p>status: {{status}}</p>
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

            const {newConnection, msg, status} = useController();
            return {msg, status};
        }
    };
</script>