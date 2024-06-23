<script>
  import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Title,
    Tooltip,
    Filler,
    Legend,
    LineController,
  } from "chart.js";
    import { onMount } from "svelte";
  // import { onMount } from "svelte";

  /**
   * @type {string}
   */
  export let project = null;

  /**
   * @type {HTMLCanvasElement}
   */
  let burndown;

  /**
   * @type {string}
   */
  let error;

  /**
   * @type {ChartJS}
   */
  let chart;

  ChartJS.register(
    CategoryScale,
    LinearScale,
    PointElement,
    LineController,
    LineElement,
    Title,
    Tooltip,
    Filler,
    Legend,
  );

  /**
   * @type {string}
   */
  const basePath = import.meta.env.VITE_API_BASE_PATH || "/api";

  onMount(async () => {
    if (!project) {
      return;
    }
    
    await plotBurnupForProject();
  });

  async function plotBurnupForProject() {
    try {
      if (chart) {
        chart.destroy();
      }

      const response = await fetch(`${basePath}/projects/${project}/burnup`);

      if (!response.ok) {
        const errorJson = await response.json();
        error = errorJson.errors.join(", ");
        return;
      }

      /**
       * @type {Array<{projectDay: string, status: string, qty: number}>}
       */
      const data = (await response.json()) || [];
      /**
       * @type {Map<string, Map<string, number>>}
       */
      const mappedData = new Map();
      const labels = [];
      const dataSets = [];

      // map the items in data into a collection for labels with unique statuses
      for (const item of data) {
        if (!mappedData.has(item.status)) {
          mappedData.set(item.status, new Map());
        }
        const projectDay = item.projectDay.substring(0, 10);
        mappedData.get(item.status).set(projectDay, item.qty);
      }

      var dynamicColors = function () {
        var r = Math.floor(Math.random() * 255);
        var g = Math.floor(Math.random() * 255);
        var b = Math.floor(Math.random() * 255);
        return "rgb(" + r + "," + g + "," + b + ")";
      };

      for (const item of mappedData) {
        const color = dynamicColors();
        const dataSet = {
          fill: true,
          label: item[0],
          data: [...item[1].values()],
          backgroundColor: color,
          borderColor: color,
        };
        dataSets.push(dataSet);

        if (labels.length === 0) {
          labels.push(...item[1].keys());
        }
      }

      const ctx = burndown.getContext("2d");
      const config = {
        type: "line",
        data: {
          labels,
          datasets: dataSets,
        },
        options: {
          responsive: true,
          plugins: {
            legend: {
              position: "top",
            },
            title: {
              display: true,
              text: "Burnup",
            },
          },
          scales: {
            y: {
              stacked: true,
            },
          },
        },
      };
      // @ts-ignore
      chart = new ChartJS(ctx, config);
    } catch (err) {
      error = err.message;
    }
  }
</script>

<canvas bind:this={burndown} width={500} height={400} />

{#if error}
  <p>{error}</p>
{/if}
