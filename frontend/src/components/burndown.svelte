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

  onMount(async () => {
    const response = await fetch(`http://localhost:8000/api/iteration/1/burndown`);
    const data = await response.json();
    const labels = [];
    const actual = [];
    const ideal = [];
    const today = new Date().toISOString().substring(0, 10);

    for (const item of data) {
      labels.push(item.IterationDay);
      actual.push(today >= item.IterationDay ? item.Remaining : null);
      ideal.push(item.Ideal);
    }

    const ctx = burndown.getContext("2d");
    const config = {
      type: "line",
      data: {
        labels,
        datasets: [
          {
            fill: true,
            label: "Actual",
            data: actual,
            borderColor: "rgb(53, 162, 235)",
            backgroundColor: "rgba(53, 162, 235, 0.3)",
          },
          {
            label: "Ideal",
            data: ideal,
            borderColor: "gray",
            backgroundColor: "gray",
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
            text: "Burndown",
          },
        },
      },
    };
    // @ts-ignore
    new ChartJS(ctx, config);
  });
</script>

<canvas bind:this={burndown} width={600} height={400} />
