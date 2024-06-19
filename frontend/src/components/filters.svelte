<script>
    import { createEventDispatcher, onMount } from "svelte";

    /**
     * @type {string}
     */
    let selectedIteration = null;

    /**
     * @type {string}
     */
    let selectedProject = null;

    /**
     * @type {{id: string, text: string, isCurrent: boolean}[]}
     */
    let iterations = [];

    /**
     * @type {{id: string, text: string}[]}
     */
    let projects = [];

    /**
     * @type {string}
     */
    let error;

    /**
     * @type {string}
     */
    const basePath = import.meta.env.VITE_API_BASE_PATH || "/api";

    const dispatch = createEventDispatcher();

    onMount(async () => {
        await loadProjects();
    });

    async function loadIterations() {
        const response = await fetch(
            `${basePath}/projects/${selectedProject}/iterations`,
        );

        if (!response.ok) {
            const errorJson = await response.json();
            error = errorJson.errors.join(", ");
            return;
        }
        iterations = [];

        const data = (await response.json()) || [];
        for (const item of data) {
            iterations = [
                ...iterations,
                {
                    id: item.id,
                    text: item.title,
                    isCurrent: isCurrentIteration(item),
                },
            ];
        }

        selectedIteration = iterations.find((i) => i.isCurrent)?.id ?? null;
        iterationChanged();
    }

    async function loadProjects() {
        const response = await fetch(`${basePath}/projects`);

        if (!response.ok) {
            const errorJson = await response.json();
            error = errorJson.errors.join(", ");
            return;
        }

        const data = (await response.json()) || [];
        for (const item of data) {
            projects = [
                ...projects,
                {
                    id: item.id,
                    text: item.title,
                },
            ];
        }

        selectedProject = projects[0].id ?? null;
        await projectChanged();
        await loadIterations();
    }

    /**
     * @param {{ title: string; startDate: string; endDate: string; }} item
     */
    function isCurrentIteration(item) {
        const today = new Date().toISOString().substring(0, 10);
        return item.startDate <= today && item.endDate >= today;
    }

    function iterationChanged() {
        dispatch("iterationChanged", { iterationId: selectedIteration });
    }
    async function projectChanged() {
        dispatch("projectChanged", { projectId: selectedProject });
        await loadIterations();
    }
</script>

<div class="filter-container">
    <select bind:value={selectedProject} on:change={projectChanged}>
        <option value={null}>Select Project...</option>
        {#each projects as project}
            <option value={project.id}>{project.text}</option>
        {/each}
    </select>

    <select bind:value={selectedIteration} on:change={iterationChanged}>
        <option value={null}>Select Iteration...</option>
        {#each iterations as iteration}
            <option value={iteration.id}>{iteration.text}</option>
        {/each}
    </select>
</div>

<style>
    select {
        padding: 6px 12px;
        font-size: 16px;
        font-weight: 400;
        line-height: 1.5;
        color: #212529;
        background-color: #fff;
        background-clip: padding-box;
        border: 1px solid #ced4da;
        border-radius: 4px;
        transition:
            border-color 0.15s ease-in-out,
            box-shadow 0.15s ease-in-out;
    }
    select:focus {
        color: #212529;
        background-color: #fff;
        border-color: #86b7fe;
        outline: 0;
        box-shadow: 0 0 0 0.25rem rgb(13 110 253 / 25%);
    }
    .filter-container {
        margin: 0 auto;
    }
</style>
