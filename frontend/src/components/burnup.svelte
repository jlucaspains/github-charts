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

  /**
   * @type {HTMLCanvasElement}
   */
  let burndown;

  /**
   * @type {string}
   */
  let error;

  /**
   * @type {string}
   */
  let selectedIteration = "1";

  /**
   * @type {{id: string, text: string, isCurrent: boolean}[]}
   */
  let iterations = [];

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

  const basePath = import.meta.env.VITE_API_BASE_PATH || "/api";

  onMount(async () => {
    await plotBurnupForProject("1");
  });

  /**
   * @param {string} project
   */
  async function plotBurnupForProject(project) {
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

      const data = (await response.json()) || [];
      const labels = [];
      const inProgress = [];
      const complete = [];

      for (const item of data) {
        labels.push(item.ProjectDay);
        inProgress.push(item.Remaining);
        complete.push(item.Done);
      }
      console.log(labels, inProgress, complete)

      const ctx = burndown.getContext("2d");
      const config = {
        type: "line",
        data: {
          labels,
          datasets: [
            {
              fill: true,
              label: "Complete",
              data: complete,
              borderColor: "rgb(53, 235, 162)",
              backgroundColor: "rgba(53, 235, 162, 0.3)",
            },
            {
              fill: true,
              label: "In Progress",
              data: inProgress,
              borderColor: "rgb(53, 162, 235)",
              backgroundColor: "rgba(53, 162, 235, 0.3)",
            },
          ],
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

<canvas bind:this={burndown} width={600} height={400} />

{#if error}
  <p>{error}</p>
{/if}
