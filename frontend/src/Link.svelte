<script>
	import { format, fromUnixTime } from 'date-fns';

	export let date;
	export let url;
	export let title;
	export let tags;
</script>

<div class="Link">
			<div class="Link-date">{format(fromUnixTime(date), 'yyyy-MM-dd')}</div>
			<h3 class="Link-title"><a href={url}>{title} &rarr;</a></h3>
			<ul class="Tags">
			{#each tags.split(',') as tag}
				<li class="Tags-tag">{tag}</li>
			{/each}
			</ul>
</div>

<style>
	.Link:not(:last-of-type) {
		margin-bottom: 6.4rem;
	}

	.Link-date {
		font-size: 1.2rem;
		margin-bottom: 1rem;
	}

	.Link-title {
		font-size: 1.8rem;
		line-height: 2.4rem;
		font-weight: 500;
		margin-bottom: 1rem;
	}

	.Link-title a {
		text-decoration: none;
		color: inherit;
	}


	.Tags {
		list-style: none;
		display: flex;
		margin: 0;
		padding: 0;
	}

	.Tags-tag {
		font-size: 1.4rem;

		background-color: #78ffd6;
		padding: 0.5rem 1rem;
		border-radius: 5px;

	
	}

	.Tags-tag:not(:last-child) {
		margin-right: 1rem;
	}
</style>