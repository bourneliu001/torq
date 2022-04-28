import "./table_controls.scss";
import {
  Navigation20Regular as NavigationIcon,
  Save20Regular as SaveIcon,
} from "@fluentui/react-icons";

import TimeIntervalSelect from "../../timeIntervalSelect/TimeIntervalSelect";
import DefaultButton from "../../buttons/Button";
import { useAppDispatch, useAppSelector } from "../../../store/hooks";
import { toggleNav } from "../../navigation/navSlice";
import SortControls from "./sort/SortControls";
import GroupPopover from "./group/GroupPopover";
import {
  selectCurrentView,
  selectedViewIndex
} from "../tableSlice";
import { useUpdateTableViewMutation, useCreateTableViewMutation } from 'apiSlice'
import FilterPopover from "./filter/FilterPopover";

import ViewsPopover from "./views/ViewsPopover";
import ColumnsPopover from "./columns/ColumnsPopover";

function TableControls() {
  const dispatch = useAppDispatch();
  const currentView = useAppSelector(selectCurrentView);
  const currentViewIndex = useAppSelector(selectedViewIndex);
  const [updateTableView] = useUpdateTableViewMutation()
  const [createTableView] = useCreateTableViewMutation()
  const saveView = () => {
    let viewMod = { ...currentView }
    viewMod.saved = true
    if (currentView.id === undefined || null) {
      createTableView({ view: viewMod, index: currentViewIndex })
      return
    }
    updateTableView(viewMod)

  }
  return (
    <div className="table-controls">
      <div className="left-container">
        <div className="upper-container">
          <DefaultButton
            icon={<NavigationIcon />}
            text={"Menu"}
            onClick={() => dispatch(toggleNav())}
            className={"show-nav-btn collapse-tablet"}
          />
          <ViewsPopover />
          {!currentView.saved && (<DefaultButton
            icon={<SaveIcon />}
            text={"Save"}
            onClick={saveView}
            className={"collapse-tablet danger"}
          />)}
        </div>
        <div className="lower-container">
          <ColumnsPopover />
          <FilterPopover />
          <SortControls />
          <GroupPopover />
        </div>
      </div>
      <div className="right-container">
        <TimeIntervalSelect />
      </div>
    </div>
  );
}

export default TableControls;