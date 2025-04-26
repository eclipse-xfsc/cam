# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Contributors:
#	Fraunhofer AISEC

// Creates a piechart with the provided props
<script lang="ts">
import { defineComponent, h, PropType, ref } from 'vue'
import { Pie } from 'vue-chartjs'
import {
  Chart as ChartJS,
  Title,
  Tooltip,
  Legend,
  ArcElement,
  CategoryScale,
  Plugin
} from 'chart.js'

ChartJS.register(Title, Tooltip, Legend, ArcElement, CategoryScale)

export default defineComponent({
  name: 'PieChart',
  components: {
    Pie
  },
  props: {
    chartId: {
      type: String,
      default: 'pie-chart'
    },
    width: {
      type: Number,
      default: 400
    },
    height: {
      type: Number,
      default: 400
    },
    cssClasses: {
      default: '',
      type: String
    },
    styles: {
      type: Object as PropType<Partial<CSSStyleDeclaration>>,
      default: () => { }
    },
    plugins: {
      type: Array as PropType<Plugin<'pie'>[]>,
      default: () => []
    },
    chartData: {
      type: Object,
      required: true
    }
  },
  emits: ["hover"],
  setup(props, { emit }) {
    const current = ref(-1);
    const chartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      onHover: function (_, item) {
        if (item.length) {
          if (current.value != item[0].index) {
            current.value = item[0].index;
            emit("hover", current.value);
          }
        } else {
          if (current.value != -1) {
            current.value = -1;
            emit("hover", current.value);
          }
        }
      }
    }

    return () =>
      h(Pie, {
        chartData: props.chartData,
        chartOptions,
        chartId: props.chartId,
        width: props.width,
        height: props.height,
        cssClasses: props.cssClasses,
        styles: props.styles,
        plugins: props.plugins
      })
  }
})
</script>
