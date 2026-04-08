export interface IMenu {
  id: string
  name: string
  openType: string
  link: string
  children?: IMenu[]
}
