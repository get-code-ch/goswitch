import { reactive, computed, toRefs } from "vue";

export default function useEventSpace() {
    const event = reactive({
        capacity: 4,
        attending: ["Tim", "Bob", "Bill"],
        spacesLeft: computed(() => {
            return event.capacity - event.attending.length;
        })
    })

    function increaseCapacity() { event.capacity++; }
    return { ...toRefs(event), increaseCapacity };

}
