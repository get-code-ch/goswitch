<template>
  <svg :width=graphProperties.width :height=graphProperties.height :viewBox="'0 0 1000 ' + graphProperties.height"
       ref="graph">
    <rect width="100%" height="100%" class="gs-background"/>
    <text x="500" y="20" dominant-baseline="middle" text-anchor="middle" fill="darkblue">{{
        graphProperties.title
      }}
    </text>
  </svg>
  <canvas :width=graphProperties.width :height=graphProperties.height ref="can"></canvas>
</template>


<script>
import drawGraph from "@/use/graph";
import {onMounted, ref} from "vue"

export default {
  name: "graph",
  props: ["gProps"],
  setup(props) {
    const graph = ref(null);
    const can = ref(null);
    let ctx = null;

    onMounted(() => {
      init(props.gProps);

      graph.value.appendChild(drawTextSVG(10, 20, 'Hello World!'));

      let max_y = Math.max.apply(Math, graphProperties.value.data.map(function (point) {
        return point.y + 20;
      }))
      //let max_x = graphProperties.value.data.length;

      let x_prec = 0;
      let y_prec = graphProperties.value.height;

      graphProperties.value.data.forEach((point, idx) => {
        let y = Math.round(graphProperties.value.height - point.y * graphProperties.value.height / max_y)
        graph.value.appendChild(drawLineSVG(x_prec, y_prec, idx, y));
        y_prec = y;
        x_prec = idx;
      })

      if (can.value.getContext) {
        ctx = can.value.getContext('2d');
        ctx.fillStyle = '#aaa';
        ctx.fillRect(5, 5, graphProperties.value.width, graphProperties.value.height);
        drawTextCanvas(ctx, 10,20, 'Hello World!');

        x_prec = 0;
        y_prec = graphProperties.value.height;
        graphProperties.value.data.forEach((point, idx) => {
          let y = Math.round(graphProperties.value.height - point.y * graphProperties.value.height / max_y)
          drawLineCanvas(ctx, x_prec, y_prec, idx, y);
          y_prec = y;
          x_prec = idx;
        })

      } else {
        //graph.value.appendChild(drawText(10, 20, 'No Context'));
      }

    });
    const {graphProperties, init, title, context, drawTextCanvas, drawLineCanvas, drawLineSVG, drawTextSVG} = drawGraph();
    return {title, graphProperties, context, graph, can};
  }
}
</script>

<style scoped>

</style>
