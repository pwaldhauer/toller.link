
<script>
	import { onMount } from 'svelte';
	import Link from './Link.svelte';

	export let name;

	const url = location.host == 'toller.link' ? '/' : 'http://localhost:8080/';

	let tokenInput;
	let query = '';

	async function  handleQuery(ev) {
		await loadLinks(ev.target.value);
	}

	let links = []


	let accessToken = '';

	async function loadLinks(query) {
		try {
			const res = await fetch(url + 'api/link?q=' + encodeURIComponent(query), {
			 headers: {
      			'Authorization': accessToken
    		},
			});

				const data = await res.json();

		if(data === null) {
			links = [];
			return;
		}

		links = data;
		} catch(err) {
			console.log(err);
			accessToken = '';
		}
		

	

	}

	
	onMount(async () => {
		accessToken = localStorage.getItem('token');
		if(!accessToken) {

		}

		loadLinks('');
	})

	function handleLogin() {
		localStorage.setItem('token', tokenInput)
		accessToken = tokenInput;

		loadLinks('');
	}

</script>

<main class="Container">

	<h1 class="Title"><a href="/">toller.link</a></h1>

	{#if !accessToken}
	<div class="Login">
		<input type="text" bind:value={tokenInput}>
		<button on:click={handleLogin}>Login!</button>
	</div>
	{:else}
	<div class="Search">
		<input type="text" placeholder="Suche…" class="Search-input" on:input={handleQuery} bind:value={query}>
	</div>

	<div class="Links">
	{#each links as link}
		<Link date={link.Date}
			  url={link.Url} 
			  title={link.Title} 
			  tags={link.Tags}
			  contextBody={link.ContextBody}
			  contextTitle={link.ContextTitle}
			></Link>
	{/each}
		
		
	</div>
	{/if}


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