<script>
  import Burndown from "./components/burndown.svelte";
  import Burnup from "./components/burnup.svelte";
  import Filters from "./components/filters.svelte";

  /**
   * @type {string}
   */
  let iteration = null;

  /**
   * @type {string}
   */
  let project = null;

  function iterationChanged(event) {
    console.log(event.detail);
    iteration = event.detail.iterationId;
  }

  function projectChanged(event) {
    console.log(event.detail);
    project = event.detail.projectId;
  }
</script>

<main>
  <header>
      <div>github-charts</div>
  </header>
  <div class="container">
    <div class="filters">
      <Filters
        on:iterationChanged={iterationChanged}
        on:projectChanged={projectChanged}
      />
    </div>
    <div class="flex-grid">
      <div class="card col">
        {#key project}
          <Burnup {project} />
        {/key}
      </div>
      <div class="card col">
        {#key iteration}
          <Burndown {project} {iteration} />
        {/key}
      </div>
    </div>
  </div>
</main>
