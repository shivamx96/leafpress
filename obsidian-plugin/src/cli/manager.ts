import { App, Notice } from "obsidian";
import { spawn } from "child_process";
import { promises as fs } from "fs";
import * as path from "path";
import * as os from "os";
import { CLIResult } from "./types";

export class BinaryManager {
  private app: App;
  private customBinaryPath: string;
  private vaultPath: string | null = null;
  private binaryPath: string | null = null;

  constructor(app: App, settings: any) {
    this.app = app;
    this.customBinaryPath = settings.customBinaryPath;
  }

  private getVaultPath(): string {
    if (this.vaultPath) return this.vaultPath;

    try {
      const adapter = this.app.vault.adapter as any;

      // Try different properties
      if (adapter.basePath && typeof adapter.basePath === 'string') {
        this.vaultPath = adapter.basePath;
      } else if (adapter.path && typeof adapter.path === 'string') {
        this.vaultPath = adapter.path;
      } else if ((adapter as any).vault?.dir) {
        this.vaultPath = (adapter as any).vault.dir;
      } else {
        // Fallback: construct from home + vault name
        const vaultName = this.app.vault.getName();
        this.vaultPath = path.join(os.homedir(), '.obsidian/vaults', vaultName);
      }

      if (!this.vaultPath || typeof this.vaultPath !== 'string') {
        throw new Error('Could not determine vault path');
      }

      console.log('[Leafpress] Vault path:', this.vaultPath);
      return this.vaultPath;
    } catch (err) {
      console.error('[Leafpress] Error getting vault path:', err);
      throw new Error('Failed to determine vault path');
    }
  }

  private getPlatformInfo(): {
    platform: string;
    arch: string;
    executable: string;
  } {
    const platform = process.platform;
    const arch = process.arch;
    let executable: string;

    if (platform === "darwin") {
      executable = arch === "arm64" ? "leafpress-macos-arm64" : "leafpress-macos-x64";
    } else if (platform === "linux") {
      executable = "leafpress-linux-x64";
    } else if (platform === "win32") {
      executable = "leafpress-win32.exe";
    } else {
      throw new Error(`Unsupported platform: ${platform}`);
    }

    return { platform, arch, executable };
  }

  private getBinaryPath(): string {
    if (this.customBinaryPath) {
      return this.customBinaryPath;
    }

    const { executable } = this.getPlatformInfo();
    return path.join(this.binaryDir, executable);
  }

  async ensureBinary(): Promise<void> {
    if (this.customBinaryPath) {
      return;
    }

    try {
      await fs.access(this.binaryPath);
    } catch {
      throw new Error(
        "Leafpress binary not found. Please install it first or set a custom binary path in settings."
      );
    }
  }

  async execCommand(args: string[]): Promise<CLIResult> {
    await this.ensureBinary();

    return new Promise((resolve) => {
      let stdout = "";
      let stderr = "";

      const child = spawn(this.binaryPath, args, {
        cwd: this.vaultPath,
      });

      child.stdout?.on("data", (data) => {
        stdout += data.toString();
      });

      child.stderr?.on("data", (data) => {
        stderr += data.toString();
      });

      child.on("close", (code) => {
        resolve({
          success: code === 0,
          stdout,
          stderr,
          code: code || 1,
        });
      });

      child.on("error", (err) => {
        resolve({
          success: false,
          stdout,
          stderr: err.message,
          code: 1,
        });
      });

      // Timeout after 5 minutes for deploy, 30s for others
      setTimeout(() => {
        child.kill();
        resolve({
          success: false,
          stdout,
          stderr: "Command timed out",
          code: -1,
        });
      }, 5 * 60 * 1000);
    });
  }
}
