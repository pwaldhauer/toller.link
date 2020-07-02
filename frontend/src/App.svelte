
<script>
	import { onMount } from 'svelte';

	import Link from './Link.svelte';
	export let name;

	let query = '';

	async function  handleQuery(ev) {
		await loadLinks(ev.target.value);
	}

	let links = []

	async function loadLinks(query) {
		const res = await fetch('http://localhost:8080/api/link?q=' + encodeURIComponent(query));
		const data = await res.json();

		if(data === null) {
			links = [];
			return;
		}

		links = data;

	}

	
	onMount(async () => {
		loadLinks('');
	})

</script>

<main class="Container">

	<h1 class="Title"><a href="/">toller.link</a></h1>

	<div class="Search">
		<input type="text" placeholder="Suche…" class="Search-input" on:input={handleQuery} bind:value={query}>
	</div>

	<div class="Links">
	{#each links as link}
		<Link date={link.Date}
			  url={link.Url} 
			  title={link.Title} 
			  tags={link.Tags}
			></Link>
	{/each}
		
		
	</div>
	


</main>

<style>
	.Container {
		width: 100%;
		max-width: 1024px;
		margin: auto;

		padding: 4rem 0;
	}

	.Title {
		display: inline-block;

		color: #fff;
		font-size: 3.2rem;
		font-weight: 300;
		margin-bottom: 4rem;

		padding: 2rem 4rem;
		background: linear-gradient(45deg, #007991, #78ffd6);

		border-radius: 15px;
	}


	.Title a {
		color: inherit;
		text-decoration: none;
	}

	.Search {
		margin-bottom: 8rem;
	}

	.Search-input {
		width: 100%;
		padding: 2rem 2rem;

		font-size: 2.4rem;


		color: #000;
		border: 1px solid #007991;
	}

	.Search-input::placeholder {
		color:#000;
		opacity: 0.4; 
		}
</style>