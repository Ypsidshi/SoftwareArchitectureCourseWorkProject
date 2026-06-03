workspace "SoftwareArchitectureCourseWorkProject" "C4 Level 1 - System Context" {
    model {
        client = person "Client" "Sanatorium services customer who registers, logs in, browses options, and makes bookings."
        staff = person "Staff" "Employee (manager/accountant/admin) who handles contracts and payment operations."

        sanatoriumPlatform = softwareSystem "Sanatorium Booking and Contracts Platform" "Platform for authentication, catalog browsing, bookings, contracts, and payments."

        client -> sanatoriumPlatform "Registration, login, browse catalog, create and manage bookings" "HTTPS/REST"
        staff -> sanatoriumPlatform "Manage contracts, create/track payments, operational actions" "HTTPS/REST"
    }

    views {
        systemContext sanatoriumPlatform "SystemContext" {
            include *
            autoLayout lr
            title "C4 Model Level 1 - System Context"
            description "System context for sanatorium booking and contracts platform: external actors and interactions."
        }

        theme default
    }
}
