import { ItemView, WorkspaceLeaf, Vault } from "obsidian";

export const VIEW_TYPE_LEAFPRESS = "leafpress-view";

export class LeafpressPanel extends ItemView {
  vault: Vault;

  constructor(leaf: WorkspaceLeaf, vault: Vault) {
    super(leaf);
    this.vault = vault;
  }

  getViewType() {
    return VIEW_TYPE_LEAFPRESS;
  }

  getDisplayText() {
    return "Leafpress";
  }

  getIcon() {
    return "leaf";
  }

  async onOpen() {
    const container = this.containerEl.children[1];
    container.empty();
    container.createEl("h2", { text: "Leafpress Status" });

    // TODO: Check if leafpress.json exists
    // TODO: Show deployment status
    // TODO: Add quick action buttons

    const placeholder = container.createEl("p");
    placeholder.setText("Loading...");
  }

  async onClose() {
    // Cleanup
  }
}
