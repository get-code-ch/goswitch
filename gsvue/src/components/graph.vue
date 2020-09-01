<template>
  <div>
    <p>Hello World!</p>
  </div>
  <svg width="100%" height="20" viewBox="0 0 1000 20">
    <rect width="100%" height="100%"  rx="5" ry="5" class="gs-background"/>
    <rect :width="10*10" height="100%"  rx="5" ry="5"  class="gs-progress"/>
    <text :x="10*5" y="50%" dominant-baseline="middle" text-anchor="middle" fill="white">10%</text>
    </svg>
</template>

<script>
export default {
name: "graph"
}
</script>

<style scoped>
.gs-background {
  background: lightgray;
}
.gs-progress {
  color: #005baa;
}
</style>

<!--

app.component('progress-bar', {
props: ['percent'],
template: '<svg width="100%" height="20" viewBox="0 0 1000 20">' +
'<rect width="100%" height="100%"  rx="5" ry="5" class="gs-background"/>' +
'<rect :width="percent*10" height="100%"  rx="5" ry="5"  class="gs-progress"/>' +
'<text :x="percent*5" y="50%" dominant-baseline="middle" text-anchor="middle" fill="white">{{ percent }}%</text>' +
'</svg>'
})

-->