(() => {
  const form = document.getElementById("resumeForm");
  const frame = document.getElementById("previewFrame");
  const statusEl = document.getElementById("status");
  const addEduBtn = document.getElementById("addEdu");
  const addExpBtn = document.getElementById("addExp");
  const eduList = document.getElementById("eduList");
  const expList = document.getElementById("expList");
  const fillExampleBtn = document.getElementById("fillExample");
  const fillExampleRuBtn = document.getElementById("fillExampleRu");
  const fillExampleEnBtn = document.getElementById("fillExampleEn");
  const downloadBtn = document.getElementById("downloadPdf");

  let previewTimer = null;
  let inFlight = 0;

  function setStatus(text) {
    statusEl.textContent = text || "";
  }

  function debouncePreview() {
    clearTimeout(previewTimer);
    previewTimer = setTimeout(updatePreview, 250);
  }

  function buildFormData() {
    return new FormData(form);
  }

  async function updatePreview() {
    const fd = buildFormData();
    const reqId = ++inFlight;
    setStatus("Обновление предпросмотра…");
    try {
      const res = await fetch("/api/preview", { method: "POST", body: fd });
      const text = await res.text();
      if (reqId !== inFlight) return;

      if (!res.ok) {
        setStatus(text || "Ошибка предпросмотра");
        return;
      }
      const data = JSON.parse(text);
      frame.srcdoc = data.html || "<html><body></body></html>";
      setStatus("");
    } catch (e) {
      if (reqId !== inFlight) return;
      setStatus("Ошибка сети");
    }
  }

  async function downloadPDF() {
    const fd = buildFormData();
    setStatus("Генерация PDF…");
    downloadBtn.disabled = true;
    try {
      const res = await fetch("/api/pdf", { method: "POST", body: fd });
      if (!res.ok) {
        const msg = await res.text();
        setStatus(msg || "Ошибка генерации PDF");
        return;
      }
      const blob = await res.blob();
      const cd = res.headers.get("content-disposition") || "";
      // Prefer RFC 5987 filename* (UTF-8) if present, fallback to filename="..."
      let filename = "resume.pdf";
      const matchStar = /filename\*\s*=\s*UTF-8''([^;]+)/i.exec(cd);
      if (matchStar && matchStar[1]) {
        try {
          filename = decodeURIComponent(matchStar[1]);
        } catch {
          filename = matchStar[1];
        }
      } else {
        const match = /filename="([^"]+)"/i.exec(cd);
        filename = match ? match[1] : filename;
      }

      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
      setStatus("");
    } catch (e) {
      setStatus("Ошибка сети");
    } finally {
      downloadBtn.disabled = false;
    }
  }

  function repeatItem(title, innerHtml, onRemove) {
    const wrap = document.createElement("div");
    wrap.className = "repeatItem";

    const head = document.createElement("div");
    head.className = "repeatItem__head";
    head.innerHTML = `<div class="repeatItem__title">${title}</div>`;

    const rm = document.createElement("button");
    rm.type = "button";
    rm.className = "btn btn--ghost btn--danger";
    rm.textContent = "Удалить";
    rm.addEventListener("click", () => {
      wrap.remove();
      onRemove && onRemove();
      debouncePreview();
    });
    head.appendChild(rm);

    const body = document.createElement("div");
    body.innerHTML = innerHtml;

    wrap.appendChild(head);
    wrap.appendChild(body);
    return wrap;
  }

  function addEducation() {
    const idx = eduList.children.length + 1;
    const item = repeatItem(
      `Запись #${idx}`,
      `
      <div class="row row--2">
        <label class="field">
          <span class="field__label">Учебное заведение</span>
          <input class="field__control" name="edu_institution" maxlength="100" />
        </label>
        <label class="field">
          <span class="field__label">Специальность</span>
          <input class="field__control" name="edu_specialty" maxlength="100" />
        </label>
      </div>
      <div class="row">
        <label class="field">
          <span class="field__label">Годы обучения</span>
          <input class="field__control" name="edu_years" maxlength="100" placeholder="2018–2022" />
        </label>
      </div>
      `
    );
    eduList.appendChild(item);
  }

  function addExperience() {
    const idx = expList.children.length + 1;
    const item = repeatItem(
      `Запись #${idx}`,
      `
      <div class="row row--2">
        <label class="field">
          <span class="field__label">Компания</span>
          <input class="field__control" name="exp_company" maxlength="100" />
        </label>
        <label class="field">
          <span class="field__label">Должность</span>
          <input class="field__control" name="exp_position" maxlength="100" />
        </label>
      </div>
      <div class="row">
        <label class="field">
          <span class="field__label">Период</span>
          <input class="field__control" name="exp_period" maxlength="100" placeholder="2023–2025" />
        </label>
      </div>
      <label class="field">
        <span class="field__label">Обязанности</span>
        <textarea class="field__control field__control--textarea" name="exp_duties" maxlength="1000"></textarea>
      </label>
      `
    );
    expList.appendChild(item);
  }

  const fallbackExample = {
    template: "modern",
    full_name: "Петрова Алина Сергеевна",
    phone: "+7 (915) 555-12-34",
    email: "alina.petrow@example.com",
    city: "Санкт‑Петербург",
    about:
      "Go‑разработчик (2+ года). Делаю HTTP‑сервисы на стандартной библиотеке, проектирую схемы данных и интеграции.\n" +
      "Люблю понятные API, читаемый код и измеримую производительность.",
    skills:
      "Go, net/http, templates\n" +
      "PostgreSQL, SQL\n" +
      "Docker, Linux\n" +
      "Git, CI\n" +
      "REST, JSON",
    additional:
      "Языки: Русский (родной), English (B2)\n" +
      "Сертификаты: Google IT Support (2024)\n" +
      "Ссылки: github.com/example | linkedin.com/in/example",
    education: [
      {
        edu_institution: "СПбПУ Петра Великого",
        edu_specialty: "Информатика и вычислительная техника",
        edu_years: "2019–2023",
      },
      {
        edu_institution: "Курсы",
        edu_specialty: "Backend на Go (практика)",
        edu_years: "2023",
      },
    ],
    experience: [
      {
        exp_company: "ООО «СеверТех»",
        exp_position: "Go Developer",
        exp_period: "Июль 2024 — н.в.",
        exp_duties:
          "— Разработала сервис генерации документов (HTML→PDF) и предпросмотр.\n" +
          "— Реализовала валидацию и обработку загрузки изображений.\n" +
          "— Оптимизировала обработку запросов и логирование ошибок.",
      },
      {
        exp_company: "Фриланс",
        exp_position: "Backend Engineer",
        exp_period: "2023 — 2024",
        exp_duties:
          "— Небольшие API для интернет‑магазинов и CRM.\n" +
          "— Интеграции со сторонними сервисами и базами данных.",
      },
    ],
  };

  async function loadExample(url) {
    try {
      const res = await fetch(url, { cache: "no-store" });
      if (!res.ok) throw new Error("bad status");
      const data = await res.json();
      return data && typeof data === "object" ? data : fallbackExample;
    } catch {
      return fallbackExample;
    }
  }

  function clearRepeatList(listEl) {
    while (listEl.firstChild) listEl.removeChild(listEl.firstChild);
  }

  function setFieldValue(name, value) {
    const el = form.querySelector(`[name="${name}"]`);
    if (!el) return;
    el.value = value ?? "";
  }

  async function fillExample(url = "/static/example.json") {
    const example = await loadExample(url);

    setFieldValue("template", example.template);
    setFieldValue("full_name", example.full_name);
    setFieldValue("phone", example.phone);
    setFieldValue("email", example.email);
    setFieldValue("city", example.city);
    setFieldValue("about", example.about);
    setFieldValue("skills", example.skills);
    setFieldValue("additional", example.additional);

    // Фото нельзя проставить программно (ограничение браузера) — пользователь выбирает файл вручную.

    clearRepeatList(eduList);
    for (const e of example.education || []) {
      addEducation();
      const last = eduList.lastElementChild;
      last.querySelector(`[name="edu_institution"]`).value = e.edu_institution;
      last.querySelector(`[name="edu_specialty"]`).value = e.edu_specialty;
      last.querySelector(`[name="edu_years"]`).value = e.edu_years;
    }

    clearRepeatList(expList);
    for (const e of example.experience || []) {
      addExperience();
      const last = expList.lastElementChild;
      last.querySelector(`[name="exp_company"]`).value = e.exp_company;
      last.querySelector(`[name="exp_position"]`).value = e.exp_position;
      last.querySelector(`[name="exp_period"]`).value = e.exp_period;
      last.querySelector(`[name="exp_duties"]`).value = e.exp_duties;
    }

    debouncePreview();
  }

  form.addEventListener("input", debouncePreview);
  form.addEventListener("change", debouncePreview);
  addEduBtn.addEventListener("click", () => {
    addEducation();
    debouncePreview();
  });
  addExpBtn.addEventListener("click", () => {
    addExperience();
    debouncePreview();
  });
  fillExampleBtn.addEventListener("click", () => fillExample("/static/example.json"));
  fillExampleRuBtn.addEventListener("click", () => fillExample("/static/example_ru_limits.json"));
  fillExampleEnBtn.addEventListener("click", () => fillExample("/static/example_en_limits.json"));
  downloadBtn.addEventListener("click", downloadPDF);

  // стартовые блоки
  addEducation();
  addExperience();
  updatePreview();
})();

