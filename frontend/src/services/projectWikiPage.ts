import {AuthenticatedHTTPFactory, getApiBaseUrl} from '@/helpers/fetcher'
import {objectToCamelCase, objectToSnakeCase} from '@/helpers/case'
import {downloadBlob} from '@/helpers/downloadBlob'
import type {IProjectWikiPage} from '@/modelTypes/IProjectWikiPage'
import type {IProjectWikiPageRevision} from '@/modelTypes/IProjectWikiPageRevision'
import type {IProjectWikiPageAttachment} from '@/modelTypes/IProjectWikiPageAttachment'

function v2Url(path: string): string {
	const v2Base = getApiBaseUrl().replace(/\/api\/v1\/$/, '/api/v2/')
	return new URL(v2Base + path, window.location.origin).toString()
}

export function parseProjectWikiPage(raw: Record<string, unknown>): IProjectWikiPage {
	const p = objectToCamelCase(raw)
	return {
		id: p.id,
		projectId: p.projectId,
		parentPageId: p.parentPageId ?? 0,
		title: p.title ?? '',
		content: p.content ?? '',
		position: p.position ?? 0,
		isHome: Boolean(p.isHome),
		createdBy: p.createdBy ? objectToCamelCase(p.createdBy) : undefined,
		updatedBy: p.updatedBy ? objectToCamelCase(p.updatedBy) : undefined,
		created: new Date(p.created),
		updated: new Date(p.updated),
	}
}

export function parseProjectWikiPageRevision(raw: Record<string, unknown>): IProjectWikiPageRevision {
	const r = objectToCamelCase(raw)
	return {
		id: r.id,
		pageId: r.pageId,
		title: r.title ?? '',
		content: r.content ?? '',
		createdById: r.createdById,
		createdBy: r.createdBy ? objectToCamelCase(r.createdBy) : undefined,
		created: new Date(r.created),
	}
}

export function parseProjectWikiPageAttachment(raw: Record<string, unknown>): IProjectWikiPageAttachment {
	const a = objectToCamelCase(raw)
	return {
		id: a.id,
		pageId: a.pageId,
		file: a.file ? objectToCamelCase(a.file) : {id: 0, name: '', size: 0, mimeType: '', created: new Date()},
		createdBy: a.createdBy ? objectToCamelCase(a.createdBy) : undefined,
		created: new Date(a.created),
	}
}

export interface WikiPageListParams {
	q?: string
	page?: number
	perPage?: number
}

export interface WikiPageListResult {
	items: IProjectWikiPage[]
	total: number
	page: number
	perPage: number
	totalPages: number
}

export interface WikiPageRevisionListResult {
	items: IProjectWikiPageRevision[]
	total: number
	page: number
	perPage: number
	totalPages: number
}

export function useProjectWikiPageService() {
	const http = AuthenticatedHTTPFactory()

	async function getAll(projectId: number, params: WikiPageListParams = {}): Promise<WikiPageListResult> {
		const {data} = await http.get(v2Url(`projects/${projectId}/wiki/pages`), {
			params: {
				q: params.q,
				page: params.page,
				per_page: params.perPage,
			},
		})
		return {
			items: (data.items ?? []).map((item: Record<string, unknown>) => parseProjectWikiPage(item)),
			total: data.total ?? 0,
			page: data.page ?? 1,
			perPage: data.per_page ?? 0,
			totalPages: data.total_pages ?? 1,
		}
	}

	async function get(projectId: number, pageId: number): Promise<IProjectWikiPage> {
		const {data} = await http.get(v2Url(`projects/${projectId}/wiki/pages/${pageId}`))
		return parseProjectWikiPage(data)
	}

	async function create(projectId: number, page: Partial<IProjectWikiPage>): Promise<IProjectWikiPage> {
		const {data} = await http.post(v2Url(`projects/${projectId}/wiki/pages`), objectToSnakeCase(page))
		return parseProjectWikiPage(data)
	}

	async function update(projectId: number, page: IProjectWikiPage): Promise<IProjectWikiPage> {
		const {data} = await http.put(v2Url(`projects/${projectId}/wiki/pages/${page.id}`), objectToSnakeCase(page))
		return parseProjectWikiPage(data)
	}

	async function remove(projectId: number, pageId: number): Promise<void> {
		await http.delete(v2Url(`projects/${projectId}/wiki/pages/${pageId}`))
	}

	async function getRevisions(projectId: number, pageId: number, params: WikiPageListParams = {}): Promise<WikiPageRevisionListResult> {
		const {data} = await http.get(v2Url(`projects/${projectId}/wiki/pages/${pageId}/revisions`), {
			params: {
				page: params.page,
				per_page: params.perPage,
			},
		})
		return {
			items: (data.items ?? []).map((item: Record<string, unknown>) => parseProjectWikiPageRevision(item)),
			total: data.total ?? 0,
			page: data.page ?? 1,
			perPage: data.per_page ?? 0,
			totalPages: data.total_pages ?? 1,
		}
	}

	async function getRevision(projectId: number, pageId: number, revisionId: number): Promise<IProjectWikiPageRevision> {
		const {data} = await http.get(v2Url(`projects/${projectId}/wiki/pages/${pageId}/revisions/${revisionId}`))
		return parseProjectWikiPageRevision(data)
	}

	async function getAttachments(projectId: number, pageId: number, params: WikiPageListParams = {}): Promise<{ items: IProjectWikiPageAttachment[]; total: number }> {
		const {data} = await http.get(v2Url(`projects/${projectId}/wiki/pages/${pageId}/attachments`), {
			params: {
				page: params.page,
				per_page: params.perPage,
			},
		})
		return {
			items: (data.items ?? []).map((item: Record<string, unknown>) => parseProjectWikiPageAttachment(item)),
			total: data.total ?? 0,
		}
	}

	async function uploadAttachments(projectId: number, pageId: number, filesToUpload: File[]): Promise<IProjectWikiPageAttachment[]> {
		const formData = new FormData()
		for (const file of filesToUpload) {
			formData.append('files', file)
		}
		const {data} = await http.post(v2Url(`projects/${projectId}/wiki/pages/${pageId}/attachments`), formData, {
			headers: {
				'Content-Type': 'multipart/form-data',
			},
		})
		return (data.success ?? []).map((item: Record<string, unknown>) => parseProjectWikiPageAttachment(item))
	}

	async function deleteAttachment(projectId: number, pageId: number, attachmentId: number): Promise<void> {
		await http.delete(v2Url(`projects/${projectId}/wiki/pages/${pageId}/attachments/${attachmentId}`))
	}

	async function downloadAttachment(projectId: number, pageId: number, attachment: IProjectWikiPageAttachment): Promise<void> {
		const res = await http.get(v2Url(`projects/${projectId}/wiki/pages/${pageId}/attachments/${attachment.id}`), {
			responseType: 'blob',
		})
		const blobUrl = URL.createObjectURL(res.data)
		downloadBlob(blobUrl, attachment.file.name)
		setTimeout(() => URL.revokeObjectURL(blobUrl), 1000)
	}

	function getAttachmentUrl(projectId: number, pageId: number, attachmentId: number): string {
		return v2Url(`projects/${projectId}/wiki/pages/${pageId}/attachments/${attachmentId}`)
	}

	return {
		getAll,
		get,
		create,
		update,
		remove,
		getRevisions,
		getRevision,
		getAttachments,
		uploadAttachments,
		deleteAttachment,
		downloadAttachment,
		getAttachmentUrl,
	}
}
