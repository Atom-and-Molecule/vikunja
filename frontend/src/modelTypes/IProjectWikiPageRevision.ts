import type {IAbstract} from './IAbstract'
import type {IProjectWikiPage} from '@/modelTypes/IProjectWikiPage'
import type {IUser} from '@/modelTypes/IUser'

export interface IProjectWikiPageRevision extends IAbstract {
	id: number
	pageId: IProjectWikiPage['id']
	title: string
	content: string
	createdById?: number
	createdBy?: IUser
	created: Date
}
