import argparse
import json
from pathlib import Path
import sys
import matplotlib.pyplot as plt

def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)

def load_results(json_path: Path):
    with open(json_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    results = data.get("results", [])
    rows = []
    for r in results:
        rows.append({
            "exp": int(r["exp"]),
            "ring_size": int(r["ring_size"]),
            "sign_avg_ms": float(r["sign_avg_ms"]),
            "verify_avg_ms": float(r["verify_avg_ms"]),
            "sig_len_avg_bytes": float(r["sig_len_avg_bytes"]),
        })
    # сортируем по размеру кольца (равносильно сортировке по exp при фиксированной базе)
    rows.sort(key=lambda x: x["ring_size"])
    return data.get("config", {}), rows

def print_summary(rows):
    # Компактная таблица в консоль (на русском)
    header = f"{'степень':>7} | {'кольцо':>7} | {'ср. подпись, мс':>16} | {'ср. проверка, мс':>18} | {'ср. длина, байт':>16}"
    sep = "-" * len(header)
    print(sep)
    print(header)
    print(sep)
    for r in rows:
        print(f"{r['exp']:>7} | {r['ring_size']:>7} | {r['sign_avg_ms']:>16.3f} | {r['verify_avg_ms']:>18.3f} | {r['sig_len_avg_bytes']:>16.0f}")
    print(sep)

def print_siglen_by_exp(rows):
    # Дополнительная сводка: длина подписи в зависимости от экспоненты
    header = f"{'степень':>7} | {'ср. длина, байт':>16}"
    sep = "-" * len(header)
    print(sep)
    print("Длина подписи по экспоненте (средние значения):")
    print(header)
    print(sep)
    for r in rows:
        print(f"{r['exp']:>7} | {r['sig_len_avg_bytes']:>16.0f}")
    print(sep)

def plot_xy(xs, ys, xlabel, ylabel, title, save_path: Path):
    plt.figure()
    plt.plot(xs, ys, marker="o")  # не задаем цвета/стили явно
    plt.title(title)
    plt.xlabel(xlabel)
    plt.ylabel(ylabel)
    plt.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()
    plt.savefig(save_path, dpi=160)
    eprint(f"Сохранено: {save_path}")

def main():
    parser = argparse.ArgumentParser(description="Построение графиков метрик Triptych по JSON.")
    parser.add_argument("--input", "-i", type=Path, default=Path("triptych_bench_results.json"),
                        help="Путь к JSON с результатами (по умолчанию: triptych_bench_results.json)")
    parser.add_argument("--outdir", "-o", type=Path, default=Path("plots"),
                        help="Каталог для сохранения графиков (по умолчанию: ./plots)")
    parser.add_argument("--no-show", action="store_true", help="Не показывать окна с графиками")
    args = parser.parse_args()

    cfg, rows = load_results(args.input)
    if not rows:
        eprint("В JSON не найдены результаты. Завершаю работу.")
        sys.exit(2)

    args.outdir.mkdir(parents=True, exist_ok=True)

    # Данные
    xs_ring = [r["ring_size"] for r in rows]
    xs_exp  = [r["exp"] for r in rows]
    sign_ms = [r["sign_avg_ms"] for r in rows]
    verify_ms = [r["verify_avg_ms"] for r in rows]
    sig_len = [r["sig_len_avg_bytes"] for r in rows]

    # Таблицы в консоль
    print_summary(rows)
    print_siglen_by_exp(rows)

    # 1) Время подписи vs размер кольца
    plot_xy(xs_ring, sign_ms,
            xlabel="Размер кольца (base^exp)",
            ylabel="Время подписи (мс)",
            title="Время подписи в зависимости от размера кольца",
            save_path=args.outdir / "sign_time_vs_ring.png")

    # 2) Время проверки vs размер кольца
    plot_xy(xs_ring, verify_ms,
            xlabel="Размер кольца (base^exp)",
            ylabel="Время проверки (мс)",
            title="Время проверки в зависимости от размера кольца",
            save_path=args.outdir / "verify_time_vs_ring.png")

    # 3) Длина подписи vs размер кольца
    plot_xy(xs_ring, sig_len,
            xlabel="Размер кольца (base^exp)",
            ylabel="Длина подписи (байт)",
            title="Длина подписи в зависимости от размера кольца",
            save_path=args.outdir / "sig_len_vs_ring.png")

    # 4) Длина подписи vs экспонента (дополнительный вывод)
    plot_xy(xs_exp, sig_len,
            xlabel="Экспонента (exp)",
            ylabel="Длина подписи (байт)",
            title="Длина подписи в зависимости от экспоненты",
            save_path=args.outdir / "sig_len_vs_exp.png")

    if not args.no_show:
        plt.show()

if __name__ == "__main__":
    main()
