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
  let selectedIteration = null;

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
    await loadIterations();
    await plotBurndownForIteration(selectedIteration);
  });

  function iterationChanged() {
    plotBurndownForIteration(selectedIteration);
  }

  /**
   * @param {string} iteration
   */
  async function plotBurndownForIteration(iteration) {
    try {
      if (chart) {
        chart.destroy();
      }

      const response = await fetch(
        `${basePath}/iterations/${iteration}/burndown`,
      );

      if (!response.ok) {
        const errorJson = await response.json();
        error = errorJson.errors.join(", ");
        return;
      }

      /**
       * @type {{ iterationDay: string; remaining: number; ideal: number; }[]}
       */
      const data = (await response.json()) || [];
      const labels = [];
      const actual = [];
      const ideal = [];
      const today = new Date().toISOString().substring(0, 10);

      for (const item of data) {
        var parsedDate = item.iterationDay.substring(0, 10);
        labels.push(parsedDate);
        actual.push(today >= parsedDate ? item.remaining : null);
        ideal.push(item.ideal);
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
              borderColor: "rgb(150, 150, 150)",
              backgroundColor: "rgb(150, 150, 150)",
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
      chart = new ChartJS(ctx, config);
    } catch (err) {
      error = err.message;
    }
  }

  async function loadIterations() {
    const response = await fetch(`${basePath}/iterations`);

    if (!response.ok) {
      const errorJson = await response.json();
      error = errorJson.errors.join(", ");
      return;
    }

    const data = (await response.json()) || [];
    for (const item of data) {
      iterations = [...iterations, { id: item.id, text: item.title, isCurrent: isCurrentIteration(item) }];
    }

    selectedIteration = iterations.find((i) => i.isCurrent)?.id ?? null;
  }

  /**
   * @param {{ title: string; startDate: string; endDate: string; }} item
   */
  function isCurrentIteration(item) {
    const today = new Date().toISOString().substring(0, 10);
    return (
      item.startDate <= today &&
      item.endDate >= today
    );
  }
</script>

<select bind:value={selectedIteration} on:change={iterationChanged}>
    <option value={null}>Select Iteration...</option>
    {#each iterations as iteration}
    <option value={iteration.id}>{iteration.text}</option>
  {/each}
</select>

<canvas bind:this={burndown} width={600} height={400} />

{#if error}
  <p>{error}</p>
{/if}
