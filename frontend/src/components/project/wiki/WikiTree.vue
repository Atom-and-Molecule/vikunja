<template>
	<aside class="menu wiki-tree">
		<div class="menu-label is-flex is-justify-content-between is-align-items-center">
			<span>{{ $t('project.wiki.pages') }}</span>
			<BaseButton
				v-if="canWrite"
				class="is-small is-primary is-light"
				:aria-label="$t('project.wiki.newPage')"
				@click="$emit('newPage')"
			>
				<span class="icon is-small">
					<Icon icon="plus" />
				</span>
				<span>{{ $t('project.wiki.newPage') }}</span>
			</BaseButton>
		</div>

		<ul class="menu-list">
			<li
				v-for="rootPage in rootPages"
				:key="rootPage.id"
			>
				<div
					class="wiki-tree-item is-flex is-align-items-center is-justify-content-between"
					:class="{'is-active': rootPage.id === currentPageId}"
					@click="$emit('selectPage', rootPage.id)"
				>
					<span class="wiki-tree-item-title is-flex is-align-items-center">
						<span
							v-if="rootPage.isHome"
							class="tag is-info is-light is-small mis-0 mie-2"
						>
							{{ $t('project.wiki.isHomeBadge') }}
						</span>
						<span class="truncate">{{ rootPage.title }}</span>
					</span>
					<BaseButton
						v-if="canWrite"
						class="is-small is-ghost p-0"
						:aria-label="$t('project.wiki.newSubpage')"
						@click.stop="$emit('newPage', rootPage.id)"
					>
						<span class="icon is-small">
							<Icon icon="plus" />
						</span>
					</BaseButton>
				</div>

				<!-- Child pages -->
				<ul
					v-if="childrenMap[rootPage.id]?.length"
					class="wiki-tree-children"
				>
					<li
						v-for="child in childrenMap[rootPage.id]"
						:key="child.id"
					>
						<div
							class="wiki-tree-item is-flex is-align-items-center is-justify-content-between"
							:class="{'is-active': child.id === currentPageId}"
							@click="$emit('selectPage', child.id)"
						>
							<span class="wiki-tree-item-title is-flex is-align-items-center">
								<span class="truncate">{{ child.title }}</span>
							</span>
							<BaseButton
								v-if="canWrite"
								class="is-small is-ghost p-0"
								:aria-label="$t('project.wiki.newSubpage')"
								@click.stop="$emit('newPage', child.id)"
							>
								<span class="icon is-small">
									<Icon icon="plus" />
								</span>
							</BaseButton>
						</div>

						<!-- Grandchild pages -->
						<ul
							v-if="childrenMap[child.id]?.length"
							class="wiki-tree-children"
						>
							<li
								v-for="grandchild in childrenMap[child.id]"
								:key="grandchild.id"
							>
								<div
									class="wiki-tree-item is-flex is-align-items-center is-justify-content-between"
									:class="{'is-active': grandchild.id === currentPageId}"
									@click="$emit('selectPage', grandchild.id)"
								>
									<span class="wiki-tree-item-title is-flex is-align-items-center">
										<span class="truncate">{{ grandchild.title }}</span>
									</span>
								</div>
							</li>
						</ul>
					</li>
				</ul>
			</li>
		</ul>

		<div
			v-if="pages.length === 0"
			class="p-4 has-text-centered has-text-grey is-size-7"
		>
			{{ $t('project.wiki.noPages') }}
		</div>
	</aside>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import type {IProjectWikiPage} from '@/modelTypes/IProjectWikiPage'
import BaseButton from '@/components/base/BaseButton.vue'

const props = defineProps<{
	pages: IProjectWikiPage[],
	currentPageId?: number,
	canWrite: boolean,
}>()

defineEmits<{
	selectPage: [pageId: number],
	newPage: [parentPageId?: number],
}>()

const rootPages = computed(() => {
	return props.pages.filter(p => !p.parentPageId || p.parentPageId === 0)
})

const childrenMap = computed(() => {
	const map: Record<number, IProjectWikiPage[]> = {}
	for (const page of props.pages) {
		if (page.parentPageId && page.parentPageId > 0) {
			if (!map[page.parentPageId]) {
				map[page.parentPageId] = []
			}
			map[page.parentPageId].push(page)
		}
	}
	return map
})
</script>

<style lang="scss" scoped>
.wiki-tree {
	background: var(--white);
	border-radius: $radius;
	box-shadow: var(--shadow-sm);
	padding: 1rem;
}

.wiki-tree-item {
	padding: 0.4rem 0.6rem;
	border-radius: $radius;
	cursor: pointer;
	transition: background-color 0.15s ease;

	&:hover {
		background-color: var(--grey-light-light);
	}

	&.is-active {
		background-color: var(--primary-light);
		color: var(--primary);
		font-weight: 600;
	}
}

.wiki-tree-item-title {
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	flex-grow: 1;
}

.wiki-tree-children {
	padding-inline-start: 1rem;
	margin-block-start: 0.25rem;
}
</style>
