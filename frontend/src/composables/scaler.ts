export class Scaler {
  private scaler: HTMLInputElement;
  private readonly scalerInitialPercentage: number;
  private readonly baseScale: number;
  private readonly scalerMinValue: number;
  private readonly scalerMaxValue: number;
  private readonly storageKey?: string;
  private readonly defaultScalerValue: number;
  private readonly initialScalerValue: number;
  private readonly hasStoredInitialValue: boolean;
  private currentFontSize: string;

  private onInput = () => this.updateScale();
  private onAuxClick = () => this.reset();

  public onUpdateScale: ((fontSize: string, rawValue: number) => void) | undefined;

  public constructor(
    scaler: HTMLInputElement,
    scalerInitialPercentage: number,
    baseScale: number,
    storageKey?: string
  ) {
    this.scaler = scaler;
    this.scalerInitialPercentage = scalerInitialPercentage;
    this.baseScale = baseScale;
    this.storageKey = storageKey;

    this.scalerMinValue = Number(scaler.min);
    this.scalerMaxValue = Number(scaler.max);
    this.defaultScalerValue = this.evalDefaultValue();
    const storedValue = this.loadStoredValue();
    this.hasStoredInitialValue = storedValue !== undefined;
    this.initialScalerValue = this.evalInitialValue(storedValue);
    this.currentFontSize = "";

    this.scaler.value = String(this.initialScalerValue);
    this.updateScale();

    this.scaler.addEventListener("input", this.onInput);
    this.scaler.addEventListener("auxclick", this.onAuxClick);
  }

  public reset(): void {
    this.scaler.value = String(this.defaultScalerValue);
    this.updateScale();
  }

  public dispose(): void {
    this.scaler.removeEventListener("input", this.onInput);
    this.scaler.removeEventListener("auxclick", this.onAuxClick);
  }

  private evalInitialValue(storedValue?: number): number {
    if (storedValue !== undefined) {
      return storedValue;
    }

    return this.defaultScalerValue;
  }

  private evalDefaultValue(): number {
    return (
      this.scalerMinValue +
      ((this.scalerMaxValue - this.scalerMinValue) * this.scalerInitialPercentage) / 100.0
    );
  }

  private loadStoredValue(): number | undefined {
    if (!this.storageKey) {
      return;
    }

    try {
      const storedValue = localStorage.getItem(this.storageKey);
      if (storedValue === null) {
        return;
      }

      const parsedValue = Number(storedValue);
      if (
        Number.isNaN(parsedValue) ||
        parsedValue < this.scalerMinValue ||
        parsedValue > this.scalerMaxValue
      ) {
        return;
      }

      return parsedValue;
    } catch {
      return;
    }
  }

  private saveCurrentValue(): void {
    if (!this.storageKey) {
      return;
    }

    try {
      localStorage.setItem(this.storageKey, String(this.currentValue()));
    } catch {
      // Ignore persistence failures (e.g. disabled localStorage).
    }
  }

  public currentValue(): number {
    return Number(this.scaler.value);
  }

  public usesStoredInitialValue(): boolean {
    return this.hasStoredInitialValue;
  }

  private evalScaler(): void {
    this.currentFontSize = Math.round(this.baseScale - this.currentValue() / 7) + "px";
  }

  public updateScale(): void {
    this.evalScaler();
    this.saveCurrentValue();

    if (this.onUpdateScale) {
      this.onUpdateScale(this.currentFontSize, this.currentValue());
    }
  }
}
