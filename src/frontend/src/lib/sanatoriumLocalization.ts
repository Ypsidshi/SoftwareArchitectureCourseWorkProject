import type { Language } from "@/i18n";
import type { Sanatorium } from "@/types";
import { displayProfile as profileLabel } from "@/lib/medicalProfiles";

const ruCity: Record<string, string> = {
  Sochi: "Сочи",
  Kislovodsk: "Кисловодск",
  Svetlogorsk: "Светлогорск",
  Gelendzhik: "Геленджик",
  Belokurikha: "Белокуриха",
  Anapa: "Анапа",
  Yessentuki: "Ессентуки",
  Yalta: "Ялта",
};

const ruAddress: Record<string, string> = {
  "Kurortny Ave, 10": "Курортный пр-т, 10",
  "Park Lane, 7": "Парковый пер., 7",
  "Morskaya St, 8": "Морская, 8",
  "Embankment, 21": "Набережная, 21",
  "Lesnaya St, 5": "Лесная, 5",
  "Pionersky Ave, 46": "Пионерский пр-т, 46",
  "Kurortnaya St, 14": "Курортная, 14",
  "Primorskaya St, 32": "Приморская, 32",
};

const ruName: Record<string, string> = {
  "Sea Breeze Health Resort": "Санаторий Морской Бриз",
  "Mountain Valley Sanatorium": "Санаторий Горная Долина",
  "Sunrise Sanatorium": "Санаторий Рассвет",
  "Azure Coast Sanatorium": "Санаторий Лазурный Берег",
  "Pine Forest Sanatorium": "Санаторий Сосновый Бор",
  "Sea Star Sanatorium": "Санаторий Морская Звезда",
  "Mountain Spring Sanatorium": "Санаторий Горный Источник",
  "Southern Terraces Sanatorium": "Санаторий Южные Террасы",
};

const ruDescription: Record<string, string> = {
  "Modern sanatorium for family and therapeutic rest.":
    "Современный санаторий для семейного и лечебного отдыха.",
  "Quiet mountain complex for respiratory and rehabilitation programs.":
    "Тихий горный комплекс для программ лечения дыхательной системы и реабилитации.",
  "Spacious seaside sanatorium with wellness and rehabilitation programs.":
    "Просторный приморский санаторий с программами оздоровления и реабилитации.",
  "Modern seaside sanatorium for wellness and recreation.":
    "Современный санаторий на побережье для отдыха и оздоровления.",
  "Forest sanatorium with rehabilitation and cardiology programs.":
    "Санаторий в лесной зоне с программами реабилитации и кардиопрофилем.",
  "Resort complex for family vacation and respiratory prevention.":
    "Курортный комплекс для семейного отдыха и профилактики дыхательной системы.",
  "Mountain sanatorium with thermal treatments and calm atmosphere.":
    "Горный санаторий с термальными процедурами и спокойной атмосферой.",
  "Southern wellness boarding house with musculoskeletal therapy programs.":
    "Южный пансионат санаторного типа с программами опорно-двигательной терапии.",
};

export {
  MEDICAL_PROFILE_SLUGS,
  displayProfile,
  profileSlugFromInput,
  profileOptions,
  formatProfilesList,
  profilesInputFromSlugs,
} from "@/lib/medicalProfiles";

const ruAmenity: Record<string, string> = {
  spa: "спа",
  pool: "бассейн",
  wifi: "wi-fi",
  medical_center: "медцентр",
  mineral_water: "минеральная вода",
  gym: "тренажерный зал",
};

const enCity: Record<string, string> = {
  Сочи: "Sochi",
  Кисловодск: "Kislovodsk",
  Светлогорск: "Svetlogorsk",
  Геленджик: "Gelendzhik",
  Белокуриха: "Belokurikha",
  Анапа: "Anapa",
  Ессентуки: "Yessentuki",
  Ялта: "Yalta",
};

const enAddress: Record<string, string> = {
  "Курортный пр-т, 10": "Kurortny Ave, 10",
  "Парковый пер., 7": "Park Lane, 7",
  "Морская, 8": "Morskaya St, 8",
  "Набережная, 21": "Embankment, 21",
  "Лесная, 5": "Lesnaya St, 5",
  "Пионерский пр-т, 46": "Pionersky Ave, 46",
  "Курортная, 14": "Kurortnaya St, 14",
  "Приморская, 32": "Primorskaya St, 32",
};

const enName: Record<string, string> = {
  "Санаторий Морской Бриз": "Sea Breeze Health Resort",
  "Санаторий Горная Долина": "Mountain Valley Sanatorium",
  "Санаторий Рассвет": "Sunrise Sanatorium",
  "Санаторий Лазурный Берег": "Azure Coast Sanatorium",
  "Санаторий Сосновый Бор": "Pine Forest Sanatorium",
  "Санаторий Морская Звезда": "Sea Star Sanatorium",
  "Санаторий Горный Источник": "Mountain Spring Sanatorium",
  "Санаторий Южные Террасы": "Southern Terraces Sanatorium",
};

const enDescription: Record<string, string> = {
  "Современный санаторий для семейного и лечебного отдыха.":
    "Modern sanatorium for family and therapeutic rest.",
  "Тихий горный комплекс для программ лечения дыхательной системы и реабилитации.":
    "Quiet mountain complex for respiratory and rehabilitation programs.",
  "Просторный приморский санаторий с программами оздоровления и реабилитации.":
    "Spacious seaside sanatorium with wellness and rehabilitation programs.",
  "Современный санаторий на побережье для отдыха и оздоровления.":
    "Modern seaside sanatorium for wellness and recreation.",
  "Санаторий в лесной зоне с программами реабилитации и кардиопрофилем.":
    "Forest sanatorium with rehabilitation and cardiology programs.",
  "Курортный комплекс для семейного отдыха и профилактики дыхательной системы.":
    "Resort complex for family vacation and respiratory prevention.",
  "Горный санаторий с термальными процедурами и спокойной атмосферой.":
    "Mountain sanatorium with thermal treatments and calm atmosphere.",
  "Южный пансионат санаторного типа с программами опорно-двигательной терапии.":
    "Southern wellness boarding house with musculoskeletal therapy programs.",
};

const enAmenity: Record<string, string> = {
  спа: "spa",
  бассейн: "pool",
  "wi-fi": "wifi",
  медцентр: "medical_center",
  "минеральная вода": "mineral_water",
  "тренажерный зал": "gym",
};

export function localizeSanatorium(item: Sanatorium, lang: Language) {
  if (lang === "en") {
    return {
      name: enName[item.name] ?? item.name,
      description: enDescription[item.description] ?? item.description,
      city: enCity[item.city] ?? item.city,
      address: enAddress[item.address] ?? item.address,
      medicalProfiles: item.medical_profiles.map((p) => profileLabel(p, "en")),
      amenities: item.amenities.map((a) => enAmenity[a] ?? a),
    };
  }

  if (lang !== "ru") {
    return {
      name: item.name,
      description: item.description,
      city: item.city,
      address: item.address,
      medicalProfiles: item.medical_profiles,
      amenities: item.amenities,
    };
  }

  return {
    name: ruName[item.name] ?? item.name,
    description: ruDescription[item.description] ?? item.description,
    city: ruCity[item.city] ?? item.city,
    address: ruAddress[item.address] ?? item.address,
    medicalProfiles: item.medical_profiles.map((p) => profileLabel(p, "ru")),
    amenities: item.amenities.map((a) => ruAmenity[a] ?? a),
  };
}
