export class Scaler {
  private scaler: HTMLInputElement;
  private readonly scalerInitialPercentage: number;
  private readonly baseScale: number;
  private readonly scalerMinValue: number;
  private readonly scalerMaxValue: number;
  // private currentImgWidth: string;
  private currentFontSize: string;

  public onUpdateScale: ((fontSize: string, rawValue: number) => void) | undefined;

  public constructor(scaler: HTMLInputElement, scalerInitialPercentage: number, baseScale: number) {
    this.scaler = scaler;
    this.scalerInitialPercentage = scalerInitialPercentage;
    this.baseScale = baseScale;

    this.scalerMinValue = Number(scaler.min);
    this.scalerMaxValue = Number(scaler.max);
    // this.currentImgWidth = "";
    this.currentFontSize = "";

    this.reset();

    this.scaler.addEventListener("input", () => this.updateScale());
    this.scaler.addEventListener("auxclick", () => this.reset());
  }

  public reset(): void {
    this.scaler.value = String(this.evalInitialValue());
    this.updateScale();
  }

  public dispose(): void {
    this.scaler.removeEventListener("input", this.updateScale);
    this.scaler.removeEventListener("auxclick", this.reset);
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
    // this.currentImgWidth = String(this.baseScale - this.currentValue()) + "rem";
    this.currentFontSize = Math.round(16 - this.currentValue() / 7) + "px";
  }

  public updateScale(): void {
    this.evalScaler();

    if (this.onUpdateScale) {
      this.onUpdateScale(this.currentFontSize, this.currentValue());
    }

    /*document.querySelectorAll(this.imgSelector).forEach((el) => {
      const img = el as HTMLElement;
      img.style.maxWidth = this.currentImgWidth;
      img.style.minWidth = this.currentImgWidth;
    });
    document
      .querySelectorAll(this.fontSelector)
      .forEach((title) => ((title as HTMLElement).style.fontSize = this.currentFontSize));*/
  }
}
