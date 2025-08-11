import { E_tagType } from "@/dataType/app.dataType";
import { E_searchLogic, E_searchSort, type I_searchData, type I_searchGroup } from "@/dataType/search.dataType";
import { defineStore } from "pinia";
export const searchStoreData = defineStore('search', {
  state: () => ({
    allId: 'all',
    allName: '全部',
    notId: 'not',
    searchData: {
      searchTextSlc: [],
      sort: E_searchSort.AddTimeDesc,
      country: {
        logic: E_searchLogic.Single,
        options: [],
      },
      definition: {
        logic: E_searchLogic.Single,
        options: [],
      },
      year: {
        logic: E_searchLogic.Single,
        options: [],
      },
      star: {
        logic: E_searchLogic.Single,
        options: [],
      },
      performer: {
        logic: E_searchLogic.Single,
        options: [],
      },
      cup: {
        logic: E_searchLogic.Single,
        options: [],
      },
      tag: {},
    } as I_searchData,
  }),
  getters: {
    tagAllOptions: (state) => {
      const opts: string[] = [];
      for (const tagClass in state.searchData.tag) {
        const tagClassData = state.searchData.tag[tagClass];
        opts.push(...tagClassData.options);
      }
      return opts;
    },
  },
  actions: {
    init: function () {
      this.clear();
    },
    clear: function () {
      this.searchData.searchTextSlc = [];
      this.searchData.country.options = [];
      this.searchData.definition.options = [];
      this.searchData.year.options = [];
      this.searchData.star.options = [];
      this.searchData.performer.options = [];
      this.searchData.cup.options = [];
      for (const key in this.searchData.tag) {
        this.searchData.tag[key].options = [];
      }
    },
    // 获取I_searchGroup
    getSearchGroup: function (type: E_tagType, diyTagClassId: string = ''): I_searchGroup | E_searchSort {
      let result: I_searchGroup | E_searchSort = this.searchData.sort;
      switch (type) {
        case E_tagType.Sort:
          result = this.searchData.sort;
          break;
        case E_tagType.Country:
          result = this.searchData.country;
          break;
        case E_tagType.Definition:
          result = this.searchData.definition;
          break;
        case E_tagType.Year:
          result = this.searchData.year;
          break;
        case E_tagType.Star:
          result = this.searchData.star;
          break;
        case E_tagType.Performer:
          result = this.searchData.performer;
          break;
        case E_tagType.Cup:
          result = this.searchData.cup;
          break;
        case E_tagType.DiyTag:
          if (diyTagClassId != '') {
            result = this.getSearchDataDiyTag(diyTagClassId);
          }
          break;
      }
      return result;
    },
    // 获取自定义标签的I_searchGroup
    getSearchDataDiyTag: function (diyTagClassId: string): I_searchGroup {
      if (!this.searchData.tag[diyTagClassId]) {
        this.searchData.tag[diyTagClassId] = {
          logic: E_searchLogic.Single,
          options: []
        }
      }
      return this.searchData.tag[diyTagClassId]
    },
    // 设置搜索逻辑
    setLogic: function (type: E_tagType, logic: E_searchLogic, diyTagClassId: string = '') {
      if (type !== E_tagType.Sort) {
        const searchGroup = this.getSearchGroup(type, diyTagClassId) as I_searchGroup;
        if (searchGroup.logic != logic) {
          searchGroup.logic = logic;
          searchGroup.options = [];
        }
      }
    },
    // 获取搜索逻辑
    getLogic: function (type: E_tagType, diyTagClassId: string = ''): E_searchLogic {
      if (type !== E_tagType.Sort) {
        const searchGroup = this.getSearchGroup(type, diyTagClassId) as I_searchGroup;
        return searchGroup.logic;
      } else {
        return E_searchLogic.Single;
      }
    },
    // 设置条件
    setQuery: function (type: E_tagType, option: string, diyTagClassId: string = '') {
      if (type == E_tagType.Sort) {
        this.searchData.sort = option as E_searchSort;
      } else {
        const searchGroup = this.getSearchGroup(type, diyTagClassId) as I_searchGroup;
        if (option == '' || option == this.allId) {
          searchGroup.options = [];
        } else if (option == this.notId) {
          searchGroup.options = [this.notId];
          searchGroup.logic = E_searchLogic.Single;
        } else {
          switch (searchGroup.logic) {
            case E_searchLogic.Single:
              const indexSingle = searchGroup.options.indexOf(option);
              if (indexSingle >= 0) {
                searchGroup.options = [];
              } else {
                searchGroup.options = [option];
              }
              break;
            case E_searchLogic.MultiAnd:
            case E_searchLogic.MultiOr:
            case E_searchLogic.Not:
              const indexOther = searchGroup.options.indexOf(option);
              if (indexOther >= 0) {
                searchGroup.options.splice(indexOther, 1);
              } else {
                searchGroup.options.push(option);
              }
              break;
          }
        }
      }
    },
    // 检查选中
    checkSelected: function (type: E_tagType, option: string, diyTagClassId: string = ''): boolean {
      let b = false;
      if (type == E_tagType.Sort) {
        b = this.searchData.sort == option as E_searchSort;
      } else {
        const searchGroup = this.getSearchGroup(type, diyTagClassId) as I_searchGroup;
        b = (option == this.allId && searchGroup.options && searchGroup.options.length == 0) || (searchGroup.options && searchGroup.options.includes(option))
      }
      return b;
    },
    // 删除搜索的diy标签
    deleteDiyTagById: function (tagId: string): boolean {
      for (const tags in this.searchData.tag) {
        const tagGroup = this.searchData.tag[tags];
        for (const index in tagGroup.options) {
          if (tagGroup.options[index] === tagId) {
            tagGroup.options.splice(parseInt(index, 10), 1);
            return true;
          }
        }
      }
      return false;
    },
  }
})
