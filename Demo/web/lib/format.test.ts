import { describe, expect, it } from "vitest";
import { formatMoney, transferTone } from "./format";

describe("financial display helpers", () => {
  it("formats exact decimal strings without number conversion", () => {
    expect(formatMoney("4820.55", "USD")).toBe("$4,820.55");
    expect(formatMoney("-42.17", "USD")).toBe("-$42.17");
  });

  it("keeps ambiguous and returned states visually distinct", () => {
    expect(transferTone("UNKNOWN")).toBe("warning");
    expect(transferTone("RETURNED")).toBe("danger");
    expect(transferTone("POSTED")).toBe("success");
  });
});
