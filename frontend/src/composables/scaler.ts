export class Scaler {
  private scaler: HTMLInputElement;
  private readonly scalerInitialPercentage: number;
  private readonly baseScale: number;
  private readonly scalerMinValue: number;
  private readonly scalerMaxValue: number;
  private currentFontSize: string;

  private onInput = () => this.updateScale();
  private onAuxClick = () => this.reset();

  public onUpdateScale: ((fontSize: string, rawValue: number) => void) | undefined;

  public constructor(scaler: HTMLInputElement, scalerInitialPercentage: number, baseScale: number) {
    this.scaler = scaler;
    this.scalerInitialPercentage = scalerInitialPercentage;
    this.baseScale = baseScale;

    this.scalerMinValue = Number(scaler.min);
    this.scalerMaxValue = Number(scaler.max);
    this.currentFontSize = "";

    this.reset();

    this.scaler.addEventListener("input", this.onInput);
    this.scaler.addEventListener("auxclick", this.onAuxClick);
  }

  public reset(): void {
    this.scaler.value = String(this.evalInitialValue());
    this.updateScale();
  }

  public dispose(): void {
    this.scaler.removeEventListener("input", this.onInput);
    this.scaler.removeEventListener("auxclick", this.onAuxClick);
  }

  private evalInitialValue(): number {
    return (
      this.scalerMinValue +
      ((this.scalerMaxValue - this.scalerMinValue) * this.scalerInitialPercentage) / 100.0
    );
  }

  public currentValue(): number {
    return Number(this.scaler.value);
  }

  private evalScaler(): void {
    this.currentFontSize = Math.round(this.baseScale - this.currentValue() / 7) + "px";
  }

  public updateScale(): void {
    this.evalScaler();

    if (this.onUpdateScale) {
      this.onUpdateScale(this.currentFontSize, this.currentValue());
    }
  }
}
