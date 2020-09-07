import {reactive, toRefs} from "vue";

export default function drawGraph() {

    const controller = reactive({
        context: null,
        graphProperties: {
            width: 1000,
            height: 400,
            title: "My Graph",
            data: []
        }
    })

    function init(props) {
        controller.graphProperties = props;
    }

    function drawTextSVG(x, y, text) {
        let t = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        t.setAttributeNS(null, 'x', x);
        t.setAttributeNS(null, 'y', y);
        t.setAttributeNS(null, 'stroke', '#0ff')
        t.appendChild(document.createTextNode(text));

        return t;
    }

    function drawLineSVG(x1, y1, x2, y2) {
        let t = document.createElementNS('http://www.w3.org/2000/svg', 'line');
        t.setAttributeNS(null, 'x1', x1);
        t.setAttributeNS(null, 'y1', y1);
        t.setAttributeNS(null, 'x2', x2);
        t.setAttributeNS(null, 'y2', y2);
        t.setAttributeNS(null, 'stroke', '#0ff')
        return t;
    }

    function drawTextCanvas(ctx, x, y, text) {
        ctx.fillStyle = '#f0f';
        ctx.fillText(text,x,y);
    }

    function drawLineCanvas(ctx, x1, y1, x2, y2) {
        ctx.fillStyle = '#f0f';
        ctx.strokeStyle = '#f0f';
        ctx.beginPath();
        ctx.moveTo(x1,y1);
        ctx.lineTo(x2,y2);
        ctx.stroke();
    }

    return {...toRefs(controller), init, drawTextSVG, drawLineSVG, drawLineCanvas, drawTextCanvas};
}