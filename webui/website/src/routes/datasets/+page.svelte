<script lang="ts">
  import Download from '@lucide/svelte/icons/download';
  import Search from '@lucide/svelte/icons/search';
  import SlidersHorizontal from '@lucide/svelte/icons/sliders-horizontal';
  import X from '@lucide/svelte/icons/x';
  import type { Dataset } from './+page.ts';

  let { data } = $props();

  const allTags = $derived([...new Set(data.datasets.flatMap((d: Dataset) => d.tags))].sort());
  const allLicenses = $derived([...new Set(data.datasets.map((d: Dataset) => d.license))].sort());
  const allAuthors = $derived([...new Set(data.datasets.map((d: Dataset) => d.author))].sort());

  let searchQuery = $state('');
  let selectedTags = $state<string[]>([]);
  let selectedLicense = $state('');
  let selectedAuthor = $state('');
  let mobileFiltersOpen = $state(false);

  function matchesQuery(d: Dataset, q: string): boolean {
    return (
      d.name.toLowerCase().includes(q) ||
      d.description.toLowerCase().includes(q) ||
      d.license.toLowerCase().includes(q) ||
      d.author.toLowerCase().includes(q) ||
      d.tags.some((t) => t.toLowerCase().includes(q))
    );
  }

  // Full filtered result
  const filtered = $derived(
    data.datasets.filter((d: Dataset) => {
      const q = searchQuery.trim().toLowerCase();
      if (q && !matchesQuery(d, q)) return false;
      if (selectedTags.length > 0 && !selectedTags.every((t) => d.tags.includes(t))) return false;
      if (selectedLicense && d.license !== selectedLicense) return false;
      if (selectedAuthor && d.author !== selectedAuthor) return false;
      return true;
    }),
  );

  // Filtered set ignoring the tag dimension — used to compute per-tag counts
  const baseWithoutTags = $derived(
    data.datasets.filter((d: Dataset) => {
      const q = searchQuery.trim().toLowerCase();
      if (q && !matchesQuery(d, q)) return false;
      if (selectedLicense && d.license !== selectedLicense) return false;
      if (selectedAuthor && d.author !== selectedAuthor) return false;
      return true;
    }),
  );

  // Filtered set ignoring the license dimension
  const baseWithoutLicense = $derived(
    data.datasets.filter((d: Dataset) => {
      const q = searchQuery.trim().toLowerCase();
      if (q && !matchesQuery(d, q)) return false;
      if (selectedTags.length > 0 && !selectedTags.every((t) => d.tags.includes(t))) return false;
      if (selectedAuthor && d.author !== selectedAuthor) return false;
      return true;
    }),
  );

  // Filtered set ignoring the author dimension
  const baseWithoutAuthor = $derived(
    data.datasets.filter((d: Dataset) => {
      const q = searchQuery.trim().toLowerCase();
      if (q && !matchesQuery(d, q)) return false;
      if (selectedTags.length > 0 && !selectedTags.every((t) => d.tags.includes(t))) return false;
      if (selectedLicense && d.license !== selectedLicense) return false;
      return true;
    }),
  );

  // How many results would remain if this tag were toggled on (or kept if already on)
  function tagCount(tag: string): number {
    const tagsToCheck = selectedTags.includes(tag) ? selectedTags : [...selectedTags, tag];
    return baseWithoutTags.filter((d) => tagsToCheck.every((t) => d.tags.includes(t))).length;
  }

  function licenseCount(license: string): number {
    return baseWithoutLicense.filter((d) => d.license === license).length;
  }

  function authorCount(author: string): number {
    return baseWithoutAuthor.filter((d) => d.author === author).length;
  }

  const hasFilters = $derived(
    searchQuery.trim() !== '' ||
      selectedTags.length > 0 ||
      selectedLicense !== '' ||
      selectedAuthor !== '',
  );

  function toggleTag(tag: string) {
    if (selectedTags.includes(tag)) {
      selectedTags = selectedTags.filter((t) => t !== tag);
    } else {
      selectedTags = [...selectedTags, tag];
    }
  }

  function clearFilters() {
    searchQuery = '';
    selectedTags = [];
    selectedLicense = '';
    selectedAuthor = '';
  }
</script>

<svelte:head>
  <title>Datasets | Hister</title>
  <meta
    name="description"
    content="Browse public datasets compatible with Hister. Filter by tag, license or author and download what you need."
  />
</svelte:head>

<div class="mx-auto md:mx-0 max-w-screen-l px-6 py-12 md:px-12">
  <!-- Page header -->
  <div class="mb-10">
    <h1
      class="font-space text-4xl font-black tracking-[-1px] text-(--text-primary) uppercase md:text-5xl"
    >
      Datasets
    </h1>
    <p class="font-inter mt-3 max-w-[70em] text-base text-(--text-secondary)">
      Public datasets you can import into Hister to extend your local index. <a
        href="https://github.com/asciimoo/hister"
        class="font-space border-brutal-border brutal-press border-[2px] bg-(--hister-cyan) px-4 py-2.5 text-[12px] font-bold tracking-[1px] text-white uppercase no-underline"
        >Submit new dataset</a
      >
    </p>
  </div>

  <!-- Sidebar + content layout -->
  <div class="flex flex-col gap-8 lg:flex-row lg:items-start lg:gap-10">
    <!-- ── Sidebar (filters) ── -->
    <aside class="lg:w-64 lg:shrink-0 xl:w-72">
      <!-- Mobile toggle -->
      <button
        onclick={() => (mobileFiltersOpen = !mobileFiltersOpen)}
        class="font-space border-brutal-border mb-4 flex w-full cursor-pointer items-center justify-between border-[3px] bg-brutal-card px-4 py-3 text-[12px] font-bold tracking-[1.5px] text-(--text-primary) uppercase lg:hidden"
      >
        <span class="flex items-center gap-2">
          <SlidersHorizontal size={14} />
          Filters
          {#if hasFilters}
            <span class="bg-(--hister-indigo) px-1.5 py-0.5 text-[10px] font-bold text-white">
              {[
                selectedTags.length > 0 ? 1 : 0,
                selectedLicense ? 1 : 0,
                selectedAuthor ? 1 : 0,
                searchQuery.trim() ? 1 : 0,
              ].reduce((a, b) => a + b, 0)}
            </span>
          {/if}
        </span>
        <X size={14} class="transition-transform {mobileFiltersOpen ? '' : 'rotate-45'}" />
      </button>

      <!-- Filter panel -->
      <div
        class="border-brutal-border flex flex-col gap-6 border-[3px] bg-brutal-card p-5 lg:sticky lg:top-6 {mobileFiltersOpen
          ? 'block'
          : 'hidden lg:flex'}"
      >
        <!-- Search -->
        <div class="flex flex-col gap-2">
          <span class="font-space font-bold tracking-[1.5px] text-(--text-secondary) uppercase"
            >Search</span
          >
          <div class="border-brutal-border relative flex items-center border-[2px]">
            <span class="pointer-events-none absolute left-3 text-(--text-secondary)">
              <Search size={14} />
            </span>
            <input
              type="text"
              placeholder="Name or description..."
              bind:value={searchQuery}
              class="font-inter w-full bg-transparent py-2 pr-8 pl-8 text-sm text-(--text-primary) outline-none placeholder:text-(--text-secondary)"
            />
            {#if searchQuery}
              <button
                onclick={() => (searchQuery = '')}
                class="absolute right-2 cursor-pointer text-(--text-secondary) hover:text-(--text-primary)"
                aria-label="Clear search"
              >
                <X size={12} />
              </button>
            {/if}
          </div>
        </div>

        <!-- Tags -->
        {#if allTags.length > 0}
          <div class="flex flex-col gap-2">
            <span class="font-space font-bold tracking-[1.5px] text-(--text-secondary) uppercase"
              >Tags</span
            >
            <ul class="m-0 flex list-none flex-col gap-1 p-0">
              {#each allTags as tag}
                {@const count = tagCount(tag)}
                {@const active = selectedTags.includes(tag)}
                <li>
                  <button
                    onclick={() => toggleTag(tag)}
                    class="font-inter flex w-full cursor-pointer items-center justify-between gap-2 px-2 py-1.5 text-sm transition-colors {active
                      ? 'bg-(--text-primary) text-white'
                      : count === 0
                        ? 'text-(--text-secondary) opacity-35'
                        : 'text-(--text-primary) hover:bg-(--muted-surface)'}"
                  >
                    <span class="truncate">{tag}</span>
                    <span
                      class="font-space shrink-0 font-bold {active
                        ? 'text-white/70'
                        : 'text-(--text-secondary)'}">{count}</span
                    >
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}

        <!-- License -->
        {#if allLicenses.length > 0}
          <div class="flex flex-col gap-2">
            <span class="font-space font-bold tracking-[1.5px] text-(--text-secondary) uppercase"
              >License</span
            >
            <ul class="m-0 flex list-none flex-col gap-1 p-0">
              {#each allLicenses as license}
                {@const count = licenseCount(license)}
                {@const active = selectedLicense === license}
                <li>
                  <button
                    onclick={() => (selectedLicense = active ? '' : license)}
                    class="font-inter flex w-full cursor-pointer items-center justify-between gap-2 px-2 py-1.5 text-sm transition-colors {active
                      ? 'bg-(--text-primary) text-white'
                      : count === 0
                        ? 'text-(--text-secondary) opacity-35'
                        : 'text-(--text-primary) hover:bg-(--muted-surface)'}"
                  >
                    <span class="truncate">{license}</span>
                    <span
                      class="font-space shrink-0 font-bold {active
                        ? 'text-white/70'
                        : 'text-(--text-secondary)'}">{count}</span
                    >
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}

        <!-- Author -->
        {#if allAuthors.length > 0}
          <div class="flex flex-col gap-2">
            <span class="font-space font-bold tracking-[1.5px] text-(--text-secondary) uppercase"
              >Author</span
            >
            <ul class="m-0 flex list-none flex-col gap-1 p-0">
              {#each allAuthors as author}
                {@const count = authorCount(author)}
                {@const active = selectedAuthor === author}
                <li>
                  <button
                    onclick={() => (selectedAuthor = active ? '' : author)}
                    class="font-inter flex w-full cursor-pointer items-center justify-between gap-2 px-2 py-1.5 text-sm transition-colors {active
                      ? 'bg-(--text-primary) text-white'
                      : count === 0
                        ? 'text-(--text-secondary) opacity-35'
                        : 'text-(--text-primary) hover:bg-(--muted-surface)'}"
                  >
                    <span class="truncate">{author}</span>
                    <span
                      class="font-space shrink-0 font-bold {active
                        ? 'text-white/70'
                        : 'text-(--text-secondary)'}">{count}</span
                    >
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}

        <!-- Clear filters -->
        {#if hasFilters}
          <button
            onclick={clearFilters}
            class="font-space flex cursor-pointer items-center justify-center gap-1.5 border-[2px] border-brutal-border px-3 py-2 text-[11px] font-semibold tracking-[0.5px] text-(--text-secondary) uppercase transition-colors hover:border-(--text-primary) hover:text-(--text-primary)"
          >
            <X size={12} />
            Clear all filters
          </button>
        {/if}
      </div>
    </aside>

    <!-- ── Main content ── -->
    <div class="min-w-0 flex-1">
      <!-- Result count bar -->
      <div class="mb-6 flex items-center justify-between gap-4">
        <p class="font-inter text-sm text-(--text-secondary)">
          <span class="font-bold text-(--text-primary)">{filtered.length}</span>
          of {data.datasets.length} dataset{data.datasets.length !== 1 ? 's' : ''}
        </p>
      </div>

      <!-- Cards grid -->
      {#if filtered.length > 0}
        <ul
          class="m-0 grid list-none gap-6 p-0 [grid-template-columns:repeat(auto-fill,minmax(min(100%,400px),1fr))]"
        >
          {#each filtered as dataset (dataset.slug)}
            <li
              class="border-brutal-border brutal-press-card flex flex-col border-[3px] bg-brutal-card"
            >
              {#if dataset.image}
                <div class="border-brutal-border overflow-hidden border-b-[3px]">
                  <img src={dataset.image} alt={dataset.name} class="h-40 w-full object-cover" />
                </div>
              {/if}

              <div class="flex flex-1 flex-col gap-3 p-6">
                <h2
                  class="font-space text-lg font-extrabold leading-tight tracking-[0.5px] text-(--text-primary)"
                >
                  {dataset.name}
                </h2>

                <div class="flex flex-wrap gap-2">
                  <span
                    class="font-space border-brutal-border border-[2px] bg-(--hister-teal) px-2 py-0.5 text-[10px] font-bold tracking-[1px] text-white uppercase"
                  >
                    {dataset.license}
                  </span>
                  <span
                    class="font-space border-brutal-border border-[2px] bg-transparent px-2 py-0.5 text-[10px] font-semibold tracking-[0.5px] text-(--text-secondary) uppercase"
                  >
                    {dataset.author}
                  </span>
                </div>

                <p class="font-inter flex-1 text-sm leading-relaxed text-(--text-secondary)">
                  {dataset.description}
                </p>

                {#if dataset.tags.length > 0}
                  <div class="flex flex-wrap gap-1.5">
                    {#each dataset.tags as tag}
                      <button
                        onclick={() => toggleTag(tag)}
                        class="font-space cursor-pointer border-[1.5px] px-2 py-0.5 text-[10px] font-semibold tracking-[0.5px] uppercase transition-all {selectedTags.includes(
                          tag,
                        )
                          ? 'border-brutal-border bg-(--text-primary) text-white'
                          : 'border-(--text-secondary) text-(--text-secondary) hover:border-(--text-primary) hover:text-(--text-primary)'}"
                      >
                        {tag}
                      </button>
                    {/each}
                  </div>
                {/if}
              </div>

              <div class="border-brutal-border border-t-[3px] p-4">
                <a
                  href={dataset.downloadUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="font-space border-brutal-border brutal-press flex w-full items-center justify-center gap-2 border-[2px] bg-(--hister-indigo) px-4 py-2.5 text-[12px] font-bold tracking-[1px] text-white uppercase no-underline"
                >
                  <Download size={14} class="shrink-0" />
                  Download
                </a>
              </div>
            </li>
          {/each}
        </ul>
      {:else}
        <div
          class="border-brutal-border flex flex-col items-center gap-4 border-[3px] bg-brutal-card px-6 py-16 text-center"
        >
          <p class="font-space text-lg font-bold text-(--text-secondary) uppercase">
            No datasets match your filters.
          </p>
          <button
            onclick={clearFilters}
            class="font-space cursor-pointer border-[2px] border-brutal-border px-5 py-2 text-[12px] font-semibold tracking-[1px] text-(--text-primary) uppercase transition-colors hover:bg-(--text-primary) hover:text-white"
          >
            Clear filters
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>
