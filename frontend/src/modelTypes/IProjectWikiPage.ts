import type {IAbstract} from './IAbstract'
import type {IProject} from '@/modelTypes/IProject'
import type {IUser} from '@/modelTypes/IUser'

export interface IProjectWikiPage extends IAbstract {
	id: number
	projectId: IProject['id']
	parentPageId: number
	title: string
	content: string
	position: number
	isHome: boolean

	createdBy?: IUser
	updatedBy?: IUser

	created: Date
	updated: Date
}
