import { App, Notice, Modal } from "obsidian";
import { BinaryManager } from "./manager";
import { CLIResult } from "./types";
import * as path from "path";

export class CommandHandlers {
  private app: App;
  private binaryManager: BinaryManager;
  private plugin: any;

  constructor(app: App, binaryManager: BinaryManager, plugin: any) {
    this.app = app;
    this.binaryManager = binaryManager;
    this.plugin = plugin;
  }

  async initialize(): Promise<void> {
    const adapter = this.app.vault.adapter as any;
    const vaultPath = adapter.basePath || this.app.vault.getName();
    const configPath = path.join(vaultPath, "leafpress.json");

    try {
      const stat = await this.app.vault.adapter.stat(configPath);
      if (stat) {
        new Notice("leafpress.json already exists");
        return;
      }
    } catch {
      // File doesn't exist, proceed
    }

    // Show initialize wizard (TODO: implement modal)
    new Notice("Initialize command placeholder");
  }

  async build(): Promise<void> {
    new Notice("Building your site...");

    const result = await this.binaryManager.execCommand(["build"]);

    if (result.success) {
      new Notice("✓ Build successful!");
    } else {
      new Notice("✗ Build failed. Check console for details.");
      console.error(result.stderr);
    }
  }

  async preview(): Promise<void> {
    new Notice("Building and previewing...");

    const result = await this.binaryManager.execCommand(["build"]);

    if (result.success) {
      new Notice("Opening preview at http://localhost:3000");
      // TODO: spawn dev server and open browser
    } else {
      new Notice("✗ Build failed");
      console.error(result.stderr);
    }
  }

  async deploy(): Promise<void> {
    new Notice("Starting deployment...");

    const result = await this.binaryManager.execCommand([
      "deploy",
      "--skip-build",
    ]);

    if (result.success) {
      // Parse deployment URL from output
      const urlMatch = result.stdout.match(/https?:\/\/[^\s]+/);
      const url = urlMatch ? urlMatch[0] : "Deployment successful";

      new Notice(`✓ Deployed: ${url}`);
      // TODO: show deployment result modal with URL
    } else {
      new Notice("✗ Deployment failed");
      console.error(result.stderr);
    }
  }
}
